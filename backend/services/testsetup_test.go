package services_test

import (
	"os"
	"path/filepath"
	"testing"

	_ "vendel/migrations"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// testDataDir holds a seeded PocketBase database shared by the DB-backed
// service tests (balance idempotency, quota reservation, terminal-refund).
// Mirrors the bootstrap pattern used by handlers/testsetup_test.go. These tests
// live in the external services_test package because vendel/migrations imports
// vendel/services — an internal test importing migrations would form a cycle.
const testDataDir = "./test_pb_data"

// testUserEmail identifies the seeded user used by DB-backed tests that rely on
// a pre-existing user (see seededUserID). Tests that need arbitrary users
// create them on demand via makeTestUser instead.
const testUserEmail = "svcuser@test.com"

// setupServicesTestApp returns a fresh TestApp backed by the seeded database.
// Each call gets an isolated copy so DB tests do not interfere with each other.
func setupServicesTestApp(t testing.TB) *tests.TestApp {
	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	return testApp
}

// newTestApp is like setupServicesTestApp but also registers app.Cleanup with
// the test, for tests that want automatic teardown.
func newTestApp(t testing.TB) *tests.TestApp {
	t.Helper()
	app, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Cleanup)
	return app
}

// seededUserID returns the id of the seeded test user.
func seededUserID(t testing.TB, app core.App) string {
	rec, err := app.FindAuthRecordByEmail("users", testUserEmail)
	if err != nil {
		t.Fatalf("failed to find seeded user: %v", err)
	}
	return rec.Id
}

// makeTestUser creates a users record and returns its id.
func makeTestUser(t testing.TB, app core.App, email string) string {
	t.Helper()
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}
	u := core.NewRecord(users)
	u.SetEmail(email)
	u.SetPassword("testpassword123")
	u.Set("full_name", "Test User")
	u.Set("is_active", true)
	if err := app.Save(u); err != nil {
		t.Fatal(err)
	}
	return u.Id
}

// makeOutgoingMessage creates an sms_messages record for the given user. status
// and retry_count let callers control terminal/retryable state; sentAt marks a
// message as already transmitted.
func makeOutgoingMessage(t testing.TB, app core.App, userId, status string, retryCount int, sentAt bool) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("sms_messages")
	if err != nil {
		t.Fatal(err)
	}
	m := core.NewRecord(col)
	m.Set("user", userId)
	m.Set("to", "+15005550006")
	m.Set("body", "hello")
	m.Set("message_type", "outgoing")
	m.Set("status", status)
	m.Set("retry_count", retryCount)
	m.Set("webhook_sent", false)
	if sentAt {
		m.Set("sent_at", "2020-01-01 00:00:00.000Z")
	}
	if err := app.Save(m); err != nil {
		t.Fatal(err)
	}
	return m
}

// quotaUsed returns the freshly-read sms_sent_this_month counter for a user.
func quotaUsed(t testing.TB, app core.App, userId string) int {
	t.Helper()
	q, err := app.FindFirstRecordByFilter("user_quotas", "user = {:u}", dbx.Params{"u": userId})
	if err != nil {
		t.Fatal(err)
	}
	return q.GetInt("sms_sent_this_month")
}

// TestMain generates the seed database if it doesn't exist, then runs tests.
func TestMain(m *testing.M) {
	if _, err := os.Stat(filepath.Join(testDataDir, "data.db")); os.IsNotExist(err) {
		if err := generateServicesTestData(); err != nil {
			panic("failed to generate services test data: " + err.Error())
		}
	}
	os.Exit(m.Run())
}

// generateServicesTestData creates a PocketBase database with all migrations
// applied plus a single active test user.
func generateServicesTestData() error {
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return err
	}

	app := core.NewBaseApp(core.BaseAppConfig{DataDir: absPath})
	if err := app.Bootstrap(); err != nil {
		return err
	}
	defer app.ResetBootstrapState()

	if err := app.RunAllMigrations(); err != nil {
		return err
	}

	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	user := core.NewRecord(users)
	user.SetEmail(testUserEmail)
	user.SetPassword("testpassword123")
	user.Set("full_name", "Service Test User")
	user.Set("is_active", true)
	return app.Save(user)
}
