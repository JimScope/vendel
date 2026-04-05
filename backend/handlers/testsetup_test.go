package handlers

import (
	"os"
	"path/filepath"
	"testing"

	_ "vendel/migrations"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const testDataDir = "./test_pb_data"

// setupTestApp creates a test PocketBase app from the seed database.
func setupTestApp(t testing.TB) *tests.TestApp {
	testApp, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}

	testApp.OnServe().BindFunc(func(se *core.ServeEvent) error {
		RegisterSMSRoutes(se)
		return se.Next()
	})

	return testApp
}

// generateTestToken creates a JWT for a test user.
func generateTestToken(t testing.TB, collectionNameOrId, email string) string {
	app, err := tests.NewTestApp(testDataDir)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	record, err := app.FindAuthRecordByEmail(collectionNameOrId, email)
	if err != nil {
		t.Fatalf("failed to find auth record for %s: %v", email, err)
	}

	token, err := record.NewAuthToken()
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// TestMain generates the seed database if it doesn't exist, then runs tests.
func TestMain(m *testing.M) {
	if _, err := os.Stat(filepath.Join(testDataDir, "data.db")); os.IsNotExist(err) {
		if err := generateTestData(); err != nil {
			panic("failed to generate test data: " + err.Error())
		}
	}
	os.Exit(m.Run())
}

// generateTestData creates a PocketBase database with migrations and seed records.
func generateTestData() error {
	absPath, err := filepath.Abs(testDataDir)
	if err != nil {
		return err
	}

	os.MkdirAll(absPath, 0755)

	// Create app directly in the test data dir (not a copy)
	app := core.NewBaseApp(core.BaseAppConfig{
		DataDir: absPath,
	})
	if err := app.Bootstrap(); err != nil {
		return err
	}
	defer app.ResetBootstrapState()

	if err := app.RunAllMigrations(); err != nil {
		return err
	}

	// Create a test superuser
	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}
	su := core.NewRecord(superusers)
	su.SetEmail("admin@test.com")
	su.SetPassword("testpassword123")
	if err := app.Save(su); err != nil {
		return err
	}

	// Create a test user
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	user := core.NewRecord(users)
	user.SetEmail("user@test.com")
	user.SetPassword("testpassword123")
	user.Set("full_name", "Test User")
	user.Set("is_active", true)
	if err := app.Save(user); err != nil {
		return err
	}

	// Create a test device
	devices, err := app.FindCollectionByNameOrId("sms_devices")
	if err != nil {
		return err
	}
	device := core.NewRecord(devices)
	device.Set("name", "Test Device")
	device.Set("phone_number", "+5355000001")
	device.Set("device_type", "android")
	device.Set("user", user.Id)
	device.Set("api_key", "test_device_key_123")
	if err := app.Save(device); err != nil {
		return err
	}

	// Create a test template
	templates, err := app.FindCollectionByNameOrId("sms_templates")
	if err != nil {
		return err
	}
	tmpl := core.NewRecord(templates)
	tmpl.Set("name", "Welcome Template")
	tmpl.Set("body", "Hello {{name}}, your code is {{code}}")
	tmpl.Set("user", user.Id)
	if err := app.Save(tmpl); err != nil {
		return err
	}

	// Create a test contact
	contacts, err := app.FindCollectionByNameOrId("contacts")
	if err != nil {
		return err
	}
	contact := core.NewRecord(contacts)
	contact.Set("name", "Alice")
	contact.Set("phone_number", "+1234567890")
	contact.Set("user", user.Id)
	if err := app.Save(contact); err != nil {
		return err
	}

	return nil
}
