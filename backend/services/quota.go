package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// QuotaError is returned when a quota limit is exceeded.
type QuotaError struct {
	StatusCode int
	Body       map[string]any
}

func (e *QuotaError) Error() string {
	b, _ := json.Marshal(e.Body)
	return string(b)
}

// GetUserQuota returns quota info for a user.
func GetUserQuota(app core.App, userId string) (map[string]any, error) {
	quota, err := getOrCreateQuota(app, userId)
	if err != nil {
		return nil, err
	}

	plan, err := app.FindRecordById("user_plans", quota.GetString("plan"))
	if err != nil {
		return nil, fmt.Errorf("plan not found")
	}

	// Calculate reset date
	var resetDate string
	lastReset := quota.GetDateTime("last_reset_date")
	if !lastReset.IsZero() {
		t := lastReset.Time()
		nextMonth := t.AddDate(0, 1, 0)
		resetDate = time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}

	return map[string]any{
		"plan":                plan.GetString("name"),
		"sms_sent_this_month": quota.GetInt("sms_sent_this_month"),
		"max_sms_per_month":   plan.GetInt("max_sms_per_month"),
		"devices_registered":  quota.GetInt("devices_registered"),
		"max_devices":         plan.GetInt("max_devices"),
		"reset_date":          resetDate,
	}, nil
}

// CheckSMSQuota verifies the user can send N SMS messages.
func CheckSMSQuota(app core.App, userId string, count int) error {
	quota, err := getOrCreateQuota(app, userId)
	if err != nil {
		return err
	}

	plan, err := app.FindRecordById("user_plans", quota.GetString("plan"))
	if err != nil {
		return fmt.Errorf("plan not found")
	}

	sent := quota.GetInt("sms_sent_this_month")
	limit := plan.GetInt("max_sms_per_month")
	available := limit - sent

	if sent+count > limit {
		return &QuotaError{
			StatusCode: 429,
			Body: map[string]any{
				"detail":      fmt.Sprintf("You can only send %d more SMS this month", available),
				"error":       "quota_exceeded",
				"quota_type":  "sms_monthly",
				"limit":       limit,
				"used":        sent,
				"available":   available,
				"upgrade_url": "/api/plans/upgrade",
			},
		}
	}

	return nil
}

// CheckDeviceQuota verifies the user can register another device.
func CheckDeviceQuota(app core.App, userId string) error {
	quota, err := getOrCreateQuota(app, userId)
	if err != nil {
		return err
	}

	plan, err := app.FindRecordById("user_plans", quota.GetString("plan"))
	if err != nil {
		return fmt.Errorf("plan not found")
	}

	registered := quota.GetInt("devices_registered")
	limit := plan.GetInt("max_devices")

	if registered >= limit {
		return &QuotaError{
			StatusCode: 429,
			Body: map[string]any{
				"detail":      fmt.Sprintf("Device limit of %d reached", limit),
				"error":       "quota_exceeded",
				"quota_type":  "devices",
				"limit":       limit,
				"used":        registered,
				"available":   0,
				"upgrade_url": "/api/plans/upgrade",
			},
		}
	}

	return nil
}

// IncrementSMSCount atomically increases the monthly SMS counter.
func IncrementSMSCount(app core.App, userId string, count int) error {
	// Ensure quota record exists
	quota, err := getOrCreateQuota(app, userId)
	if err != nil {
		return err
	}

	_, err = app.DB().
		NewQuery("UPDATE user_quotas SET sms_sent_this_month = sms_sent_this_month + {:count} WHERE id = {:id}").
		Bind(dbx.Params{"count": count, "id": quota.Id}).
		Execute()
	return err
}

// IncrementDeviceCount atomically increases the device counter.
func IncrementDeviceCount(app core.App, userId string) error {
	quota, err := getOrCreateQuota(app, userId)
	if err != nil {
		return err
	}

	_, err = app.DB().
		NewQuery("UPDATE user_quotas SET devices_registered = devices_registered + 1 WHERE id = {:id}").
		Bind(dbx.Params{"id": quota.Id}).
		Execute()
	return err
}

// DecrementDeviceCount atomically decreases the device counter.
func DecrementDeviceCount(app core.App, userId string) error {
	quota, err := getOrCreateQuota(app, userId)
	if err != nil {
		return err
	}

	_, err = app.DB().
		NewQuery("UPDATE user_quotas SET devices_registered = MAX(devices_registered - 1, 0) WHERE id = {:id}").
		Bind(dbx.Params{"id": quota.Id}).
		Execute()
	return err
}

// CreateDefaultQuota creates a quota record for a new user with the free plan.
func CreateDefaultQuota(app core.App, userId string) error {
	_, err := getOrCreateQuota(app, userId)
	return err
}

// ResetMonthlyQuotas resets all SMS counters (called by cron on the 1st of each month).
func ResetMonthlyQuotas(app core.App) error {
	records, err := app.FindRecordsByFilter("user_quotas", "1=1", "", 0, 0)
	if err != nil {
		return err
	}

	resetCount := 0
	for _, q := range records {
		q.Set("sms_sent_this_month", 0)
		q.Set("last_reset_date", time.Now().UTC().Format(time.RFC3339))
		if err := app.Save(q); err == nil {
			resetCount++
		}
	}

	log.Printf("Reset monthly quotas for %d users", resetCount)
	return nil
}

// getOrCreateQuota finds or creates a quota record for the user.
func getOrCreateQuota(app core.App, userId string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"user_quotas",
		"user = {:userId}",
		"", 1, 0,
		dbx.Params{"userId": userId},
	)
	if err == nil && len(records) > 0 {
		return records[0], nil
	}

	// Find or create free plan
	freePlan, err := findFreePlan(app)
	if err != nil {
		return nil, err
	}

	collection, err := app.FindCollectionByNameOrId("user_quotas")
	if err != nil {
		return nil, err
	}

	quota := core.NewRecord(collection)
	quota.Set("user", userId)
	quota.Set("plan", freePlan.Id)
	quota.Set("sms_sent_this_month", 0)
	quota.Set("devices_registered", 0)
	quota.Set("last_reset_date", time.Now().UTC().Format(time.RFC3339))

	if err := app.Save(quota); err != nil {
		return nil, err
	}

	return quota, nil
}

func findFreePlan(app core.App) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"user_plans",
		"name ~ 'free'",
		"", 1, 0,
	)
	if err == nil && len(records) > 0 {
		return records[0], nil
	}

	// Create free plan if it doesn't exist
	collection, err := app.FindCollectionByNameOrId("user_plans")
	if err != nil {
		return nil, err
	}

	plan := core.NewRecord(collection)
	plan.Set("name", "Free")
	plan.Set("max_sms_per_month", 50)
	plan.Set("max_devices", 1)
	plan.Set("price", 0)
	plan.Set("price_yearly", 0)
	plan.Set("is_public", true)

	if err := app.Save(plan); err != nil {
		return nil, err
	}

	return plan, nil
}

// GenerateSecureKey generates a cryptographically secure random key with a prefix.
func GenerateSecureKey(prefix string, length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	return prefix + encoded, nil
}

// containsEvent checks if a JSON array string contains a specific event.
func containsEvent(eventsJSON string, event string) bool {
	if eventsJSON == "" {
		return false
	}
	var events []string
	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil {
		return strings.Contains(eventsJSON, event)
	}
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}
