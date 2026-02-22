package main

import (
	"ender/handlers"
	"ender/middleware"
	"ender/services"
	"log"
	"os"

	_ "ender/migrations"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	// Load .env file (from cwd or parent dir)
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	app := pocketbase.New()

	// Enable auto-migration in dev mode
	isGoRun := os.Getenv("ENVIRONMENT") != "production"
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	// ── Bootstrap services on serve ──────────────────────────────────
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Initialize FCM
		services.InitFCM()

		// Seed first superuser from env vars
		seedSuperuser(se.App)

		// ── Custom API routes ────────────────────────────────────────
		handlers.RegisterSMSRoutes(se)
		handlers.RegisterPlanRoutes(se)
		handlers.RegisterWebhookRoutes(se)
		handlers.RegisterUtilRoutes(se)

		// ── Maintenance mode middleware ───────────────────────────────
		se.Router.BindFunc(middleware.MaintenanceMiddleware(se.App))

		return se.Next()
	})

	// ── Record hooks ─────────────────────────────────────────────────

	// Users: create default quota on registration
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return services.CreateDefaultQuota(e.App, e.Record.Id)
	})

	// SMS Devices: check quota + generate API key on create
	app.OnRecordCreate("sms_devices").BindFunc(func(e *core.RecordEvent) error {
		userId := e.Record.GetString("user")
		if err := services.CheckDeviceQuota(e.App, userId); err != nil {
			return err
		}

		// Generate secure API key
		apiKey, err := services.GenerateSecureKey("dk_", 32)
		if err != nil {
			return err
		}
		e.Record.Set("api_key", apiKey)

		if err := e.Next(); err != nil {
			return err
		}

		// Increment device count after successful creation
		return services.IncrementDeviceCount(e.App, userId)
	})

	// SMS Devices: decrement count on delete
	app.OnRecordDelete("sms_devices").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		userId := e.Record.GetString("user")
		return services.DecrementDeviceCount(e.App, userId)
	})

	// API Keys: generate secure key on create
	app.OnRecordCreate("api_keys").BindFunc(func(e *core.RecordEvent) error {
		key, err := services.GenerateSecureKey("ek_", 32)
		if err != nil {
			return err
		}
		e.Record.Set("key", key)
		return e.Next()
	})

	// ── Cron jobs ────────────────────────────────────────────────────
	app.Cron().MustAdd("monthly-quota-reset", "0 0 1 * *", func() {
		if err := services.ResetMonthlyQuotas(app); err != nil {
			log.Printf("ERROR: monthly quota reset failed: %v", err)
		}
	})

	app.Cron().MustAdd("daily-renewal-check", "0 8 * * *", func() {
		if err := services.CheckRenewals(app); err != nil {
			log.Printf("ERROR: renewal check failed: %v", err)
		}
	})

	app.Cron().MustAdd("retry-failed-sms", "0 */6 * * *", func() {
		if err := services.RetryFailedMessages(app); err != nil {
			log.Printf("ERROR: retry failed SMS failed: %v", err)
		}
	})

	// ── Start ────────────────────────────────────────────────────────
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// seedSuperuser creates the first superuser from environment variables.
func seedSuperuser(app core.App) {
	email := os.Getenv("FIRST_SUPERUSER")
	password := os.Getenv("FIRST_SUPERUSER_PASSWORD")
	if email == "" || password == "" {
		return
	}

	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		log.Printf("WARNING: could not find superusers collection: %v", err)
		return
	}

	// Check if already exists
	existing, _ := app.FindAuthRecordByEmail(superusers, email)
	if existing != nil {
		return
	}

	record := core.NewRecord(superusers)
	record.SetEmail(email)
	record.SetPassword(password)

	if err := app.Save(record); err != nil {
		log.Printf("WARNING: could not create superuser: %v", err)
	} else {
		log.Printf("Created superuser: %s", email)
	}
}
