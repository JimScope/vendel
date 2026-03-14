package main

import (
	"vendel/cronjobs"
	"vendel/handlers"
	"vendel/hooks"
	"vendel/middleware"
	"vendel/services"
	"fmt"
	"log"
	"os"

	_ "vendel/migrations"

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
		services.InitFCM(se.App)
		seedSuperuser(se.App)
		configureOAuth(se.App)
		configureSMTP(se.App)
		configureAuthSecurity(se.App)
		configureRateLimits(se.App)

		if os.Getenv("WEBHOOK_ENCRYPTION_KEY") == "" {
			return fmt.Errorf("WEBHOOK_ENCRYPTION_KEY environment variable is required")
		}

		// Custom API routes
		handlers.RegisterSMSRoutes(se)
		handlers.RegisterPlanRoutes(se)
		handlers.RegisterUserWebhookRoutes(se)
		handlers.RegisterWebhookRoutes(se)
		handlers.RegisterApiKeyRoutes(se)
		handlers.RegisterUtilRoutes(se)

		// Global middleware
		se.Router.BindFunc(middleware.SecurityHeadersMiddleware)
		se.Router.BindFunc(middleware.MaintenanceMiddleware(se.App))

		return se.Next()
	})

	// ── Hooks ────────────────────────────────────────────────────────
	hooks.RegisterAuthHooks(app)
	hooks.RegisterUserHooks(app)
	hooks.RegisterDeviceHooks(app)
	hooks.RegisterWebhookHooks(app)
	hooks.RegisterApiKeyHooks(app)
	hooks.RegisterScheduledSMSHooks(app)
	hooks.RegisterEncryptionHooks(app)
	hooks.RegisterRealtimeHooks(app)

	// ── Cron jobs ────────────────────────────────────────────────────
	cronjobs.RegisterCronJobs(app)

	// ── Start ────────────────────────────────────────────────────────
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
