package services_test

import (
	"os"
	"path/filepath"
	"testing"

	_ "vendel/migrations"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

// testDataDir holds a seeded PocketBase database shared by the DB-backed
// service tests (balance idempotency, quota reservation). Mirrors the
// bootstrap pattern used by handlers/testsetup_test.go. These tests live in
// the external services_test package because vendel/migrations imports
// vendel/services — an internal test importing migrations would form a cycle.
const testDataDir = "./test_pb_data"

// testUserEmail identifies the seeded user used by DB-backed tests.
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

// seededUserID returns the id of the seeded test user.
func seededUserID(t testing.TB, app core.App) string {
	rec, err := app.FindAuthRecordByEmail("users", testUserEmail)
	if err != nil {
		t.Fatalf("failed to find seeded user: %v", err)
	}
	return rec.Id
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
