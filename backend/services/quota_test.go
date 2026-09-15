package services_test

import (
	"errors"
	"sync"
	"testing"

	"vendel/services"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// setQuotaLimit points the user's quota at a plan with the given monthly SMS
// limit and resets the sent counter to zero. Returns the quota record id.
func setQuotaLimit(t testing.TB, app core.App, userID string, limit int) string {
	t.Helper()

	// Ensure the quota row exists.
	if err := services.CreateDefaultQuota(app, userID); err != nil {
		t.Fatalf("CreateDefaultQuota failed: %v", err)
	}

	// Create a dedicated plan carrying the desired limit.
	plans, err := app.FindCollectionByNameOrId("user_plans")
	if err != nil {
		t.Fatalf("user_plans collection not found: %v", err)
	}
	plan := core.NewRecord(plans)
	plan.Set("name", "TestLimitPlan")
	plan.Set("max_sms_per_month", limit)
	plan.Set("max_devices", 1)
	plan.Set("max_scheduled_sms", 1)
	plan.Set("max_integrations", 1)
	plan.Set("price", 0)
	plan.Set("price_yearly", 0)
	plan.Set("is_public", false)
	if err := app.Save(plan); err != nil {
		t.Fatalf("failed to save test plan: %v", err)
	}

	quota, err := app.FindFirstRecordByFilter("user_quotas", "user = {:u}", dbx.Params{"u": userID})
	if err != nil {
		t.Fatalf("quota not found: %v", err)
	}
	quota.Set("plan", plan.Id)
	quota.Set("sms_sent_this_month", 0)
	if err := app.Save(quota); err != nil {
		t.Fatalf("failed to update quota: %v", err)
	}
	return quota.Id
}

func quotaSent(t testing.TB, app core.App, quotaID string) int {
	t.Helper()
	rec, err := app.FindRecordById("user_quotas", quotaID)
	if err != nil {
		t.Fatalf("failed to reload quota: %v", err)
	}
	return rec.GetInt("sms_sent_this_month")
}

// TestReserveSMSQuotaConcurrent is the core over-send guard: N goroutines each
// try to reserve 1 SMS against a limit L; no more than L reservations may
// succeed and the persisted counter must equal the number of successes.
func TestReserveSMSQuotaConcurrent(t *testing.T) {
	app := setupServicesTestApp(t)
	defer app.Cleanup()

	userID := seededUserID(t, app)

	const limit = 20
	const attempts = 100
	quotaID := setQuotaLimit(t, app, userID, limit)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		succeeded int
		quotaErrs int
	)

	wg.Add(attempts)
	for i := 0; i < attempts; i++ {
		go func() {
			defer wg.Done()
			err := services.ReserveSMSQuota(app, userID, 1)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				succeeded++
				return
			}
			var qe *services.QuotaError
			if errors.As(err, &qe) {
				quotaErrs++
			} else {
				// SQLITE_BUSY and friends are acceptable under contention; they
				// simply mean the reservation did not happen. What must never
				// happen is a success beyond the limit.
				t.Errorf("unexpected non-quota error: %v", err)
			}
		}()
	}
	wg.Wait()

	if succeeded > limit {
		t.Fatalf("over-send: %d reservations succeeded, limit was %d", succeeded, limit)
	}

	// The persisted counter must exactly match the successful reservations —
	// no lost or phantom increments.
	if got := quotaSent(t, app, quotaID); got != succeeded {
		t.Fatalf("counter = %d, but %d reservations succeeded", got, succeeded)
	}

	// With attempts well above the limit, the guard should ideally let the full
	// limit through; under SQLite write contention fewer may pass, which is
	// still safe (never more than the limit).
	if succeeded != limit {
		t.Logf("succeeded=%d limit=%d (under contention fewer than limit may pass)", succeeded, limit)
	}
}

// TestReserveReleaseAccounting is a deterministic (non-concurrent) check of the
// reserve/release bookkeeping and the boundary behaviour at the limit.
func TestReserveReleaseAccounting(t *testing.T) {
	app := setupServicesTestApp(t)
	defer app.Cleanup()

	userID := seededUserID(t, app)
	const limit = 5
	quotaID := setQuotaLimit(t, app, userID, limit)

	// Reserve up to the limit in mixed batch sizes.
	if err := services.ReserveSMSQuota(app, userID, 3); err != nil {
		t.Fatalf("reserve 3 failed: %v", err)
	}
	if err := services.ReserveSMSQuota(app, userID, 2); err != nil {
		t.Fatalf("reserve 2 failed: %v", err)
	}
	if got := quotaSent(t, app, quotaID); got != 5 {
		t.Fatalf("counter = %d, want 5", got)
	}

	// One more must be rejected with a QuotaError.
	err := services.ReserveSMSQuota(app, userID, 1)
	var qe *services.QuotaError
	if !errors.As(err, &qe) {
		t.Fatalf("expected QuotaError at limit, got %v", err)
	}
	if qe.StatusCode != 429 {
		t.Errorf("QuotaError status = %d, want 429", qe.StatusCode)
	}

	// Releasing frees capacity for a subsequent reservation.
	if err := services.ReleaseSMSQuota(app, userID, 2); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	if got := quotaSent(t, app, quotaID); got != 3 {
		t.Fatalf("counter after release = %d, want 3", got)
	}
	if err := services.ReserveSMSQuota(app, userID, 2); err != nil {
		t.Fatalf("reserve after release failed: %v", err)
	}
	if got := quotaSent(t, app, quotaID); got != 5 {
		t.Fatalf("counter = %d, want 5", got)
	}

	// A batch that does not fit is rejected atomically (counter unchanged).
	if err := services.ReserveSMSQuota(app, userID, 3); !errors.As(err, &qe) {
		t.Fatalf("expected QuotaError for oversized batch, got %v", err)
	}
	if got := quotaSent(t, app, quotaID); got != 5 {
		t.Fatalf("counter after rejected batch = %d, want 5 (partial write?)", got)
	}

	// Release never drives the counter below zero.
	if err := services.ReleaseSMSQuota(app, userID, 100); err != nil {
		t.Fatalf("over-release failed: %v", err)
	}
	if got := quotaSent(t, app, quotaID); got != 0 {
		t.Fatalf("counter after over-release = %d, want 0", got)
	}
}
