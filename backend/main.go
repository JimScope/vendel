package main

import (
	"vendel/handlers"
	"vendel/middleware"
	"vendel/services"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	_ "vendel/migrations"

	"github.com/joho/godotenv"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/routine"
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
		services.InitFCM(se.App)

		// Seed first superuser from env vars
		seedSuperuser(se.App)

		// Configure OAuth2 providers from env vars
		configureOAuth(se.App)

		// Configure SMTP from env vars
		configureSMTP(se.App)

		// Configure rate limits
		configureRateLimits(se.App)

		// Require webhook encryption key
		if os.Getenv("WEBHOOK_ENCRYPTION_KEY") == "" {
			return fmt.Errorf("WEBHOOK_ENCRYPTION_KEY environment variable is required")
		}

		// ── Custom API routes ────────────────────────────────────────
		handlers.RegisterSMSRoutes(se)
		handlers.RegisterPlanRoutes(se)
		handlers.RegisterUserWebhookRoutes(se)
		handlers.RegisterWebhookRoutes(se)
		handlers.RegisterApiKeyRoutes(se)
		handlers.RegisterUtilRoutes(se)

		// ── Maintenance mode middleware ───────────────────────────────
		se.Router.BindFunc(middleware.MaintenanceMiddleware(se.App))

		return se.Next()
	})

	// ── Auth hooks ───────────────────────────────────────────────────

	// Block login for unverified users (password auth only; OAuth2 auto-verifies)
	app.OnRecordAuthWithPasswordRequest("users").BindFunc(func(e *core.RecordAuthWithPasswordRequestEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		if !e.Record.Verified() {
			return e.UnauthorizedError("Please verify your email address before logging in.", nil)
		}
		return nil
	})

	// ── Record hooks ─────────────────────────────────────────────────

	// System config: sync app_name to PocketBase settings
	app.OnRecordAfterUpdateSuccess("system_config").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("key") == "app_name" {
			settings := e.App.Settings()
			settings.Meta.AppName = e.Record.GetString("value")
			settings.Meta.SenderName = e.Record.GetString("value")
			if err := e.App.Save(settings); err != nil {
				e.App.Logger().Warn("could not sync app_name to PocketBase settings", slog.Any("error", err))
			}
		}
		return e.Next()
	})

	// Users: create default quota on registration
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		return services.CreateDefaultQuota(e.App, e.Record.Id)
	})

	// SMS Devices: check quota + generate API key + default device_type on create
	app.OnRecordCreate("sms_devices").BindFunc(func(e *core.RecordEvent) error {
		userId := e.Record.GetString("user")
		if err := services.CheckDeviceQuota(e.App, userId); err != nil {
			return err
		}

		// Increment device count before persisting so quota stays consistent
		if err := services.IncrementDeviceCount(e.App, userId); err != nil {
			return err
		}

		// Generate secure API key
		e.Record.Set("api_key", services.GenerateSecureKey("dk_", 32))

		// Default device_type to "android" if not set
		if e.Record.GetString("device_type") == "" {
			e.Record.Set("device_type", "android")
		}

		// Unhide so the key is returned in the create response (only shown once)
		e.Record.Unhide("api_key")

		if err := e.Next(); err != nil {
			// Rollback the increment if record creation fails
			_ = services.DecrementDeviceCount(e.App, userId)
			return err
		}

		return nil
	})

	// SMS Devices: decrement count on delete
	app.OnRecordDelete("sms_devices").BindFunc(func(e *core.RecordEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		userId := e.Record.GetString("user")
		return services.DecrementDeviceCount(e.App, userId)
	})

	// Webhooks: validate URL + encrypt secret_key on create/update
	validateAndEncryptWebhook := func(e *core.RecordEvent) error {
		if url := e.Record.GetString("url"); url != "" {
			if err := services.ValidateWebhookURL(url); err != nil {
				return fmt.Errorf("invalid webhook URL: %w", err)
			}
		}
		if secret := e.Record.GetString("secret_key"); secret != "" {
			encrypted, err := services.EncryptSecret(secret)
			if err != nil {
				return err
			}
			e.Record.Set("secret_key", encrypted)
		}
		return e.Next()
	}
	app.OnRecordCreate("webhook_configs").BindFunc(validateAndEncryptWebhook)
	app.OnRecordUpdate("webhook_configs").BindFunc(validateAndEncryptWebhook)

	// API Keys: generate secure key on create
	app.OnRecordCreate("api_keys").BindFunc(func(e *core.RecordEvent) error {
		e.Record.Set("key", services.GenerateSecureKey("vk_", 32))

		// Unhide so the key is returned in the create response (only shown once)
		e.Record.Unhide("key")

		return e.Next()
	})

	// Scheduled SMS: compute next_run_at on create
	app.OnRecordCreate("scheduled_sms").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("timezone") == "" {
			e.Record.Set("timezone", "UTC")
		}
		if e.Record.GetString("status") == "" {
			e.Record.Set("status", "active")
		}

		scheduleType := e.Record.GetString("schedule_type")
		if scheduleType == "one_time" {
			e.Record.Set("next_run_at", e.Record.GetString("scheduled_at"))
		} else if scheduleType == "recurring" {
			cronExpr := e.Record.GetString("cron_expression")
			tz := e.Record.GetString("timezone")
			nextRun, err := services.ComputeNextRun(cronExpr, tz)
			if err != nil {
				return fmt.Errorf("invalid cron expression: %w", err)
			}
			e.Record.Set("next_run_at", nextRun)
		}

		return e.Next()
	})

	// Scheduled SMS: recompute next_run_at on update
	app.OnRecordUpdate("scheduled_sms").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("timezone") == "" {
			e.Record.Set("timezone", "UTC")
		}

		scheduleType := e.Record.GetString("schedule_type")
		if scheduleType == "one_time" {
			e.Record.Set("next_run_at", e.Record.GetString("scheduled_at"))
		} else if scheduleType == "recurring" {
			cronExpr := e.Record.GetString("cron_expression")
			tz := e.Record.GetString("timezone")
			nextRun, err := services.ComputeNextRun(cronExpr, tz)
			if err != nil {
				return fmt.Errorf("invalid cron expression: %w", err)
			}
			e.Record.Set("next_run_at", nextRun)
		}

		return e.Next()
	})

	// SMS Messages: notify modem agents via SSE when messages are assigned
	notifyModemIfAssigned := func(e *core.RecordEvent) error {
		if e.Record.GetString("status") != "assigned" {
			return e.Next()
		}
		deviceId := e.Record.GetString("device")
		if deviceId == "" {
			return e.Next()
		}
		device, err := e.App.FindRecordById("sms_devices", deviceId)
		if err != nil || device.GetString("device_type") != "modem" {
			return e.Next()
		}
		routine.FireAndForget(func() { services.NotifyModemAgent(e.App, deviceId, e.Record) })
		return e.Next()
	}
	app.OnRecordAfterCreateSuccess("sms_messages").BindFunc(notifyModemIfAssigned)
	app.OnRecordAfterUpdateSuccess("sms_messages").BindFunc(notifyModemIfAssigned)

	// Realtime: guard modem/* subscriptions and broadcast status on relevant subscriptions
	app.OnRealtimeSubscribeRequest().BindFunc(func(e *core.RealtimeSubscribeRequestEvent) error {
		hasModemSub := false
		hasStatusSub := false
		for _, sub := range e.Subscriptions {
			if sub == "modem-status" {
				hasStatusSub = true
				continue
			}
			if !strings.HasPrefix(sub, "modem/") {
				continue
			}
			hasModemSub = true
			deviceId := strings.TrimPrefix(sub, "modem/")
			apiKey := e.Request.Header.Get("X-API-Key")
			if apiKey == "" {
				return fmt.Errorf("authentication required for modem subscriptions")
			}
			_, err := e.App.FindFirstRecordByFilter(
				"sms_devices",
				"id = {:id} && api_key = {:key}",
				dbx.Params{"id": deviceId, "key": apiKey},
			)
			if err != nil {
				return fmt.Errorf("unauthorized modem subscription")
			}
		}
		if err := e.Next(); err != nil {
			return err
		}
		// Broadcast modem status when an agent connects or a frontend subscribes
		if hasModemSub || hasStatusSub {
			routine.FireAndForget(func() { services.BroadcastModemStatus(e.App) })
		}
		return nil
	})

	// Realtime: broadcast modem status when any SSE client disconnects
	app.OnRealtimeConnectRequest().BindFunc(func(e *core.RealtimeConnectRequestEvent) error {
		// e.Next() blocks until the client disconnects
		if err := e.Next(); err != nil {
			return err
		}
		// Client disconnected — broadcast updated modem status
		routine.FireAndForget(func() { services.BroadcastModemStatus(e.App) })
		return nil
	})

	// ── Cron jobs ────────────────────────────────────────────────────
	app.Cron().MustAdd("monthly-quota-reset", "0 0 1 * *", func() {
		if err := services.ResetMonthlyQuotas(app); err != nil {
			app.Logger().Error("monthly quota reset failed", slog.Any("error", err))
		}
	})

	app.Cron().MustAdd("daily-renewal-check", "0 8 * * *", func() {
		if err := services.CheckRenewals(app); err != nil {
			app.Logger().Error("renewal check failed", slog.Any("error", err))
		}
	})

	app.Cron().MustAdd("retry-failed-sms", "*/15 * * * *", func() {
		if err := services.RetryFailedMessages(app); err != nil {
			app.Logger().Error("retry failed SMS", slog.Any("error", err))
		}
	})

	app.Cron().MustAdd("retry-failed-webhooks", "*/1 * * * *", func() {
		if err := services.RetryFailedWebhooks(app); err != nil {
			app.Logger().Error("retry failed webhooks", slog.Any("error", err))
		}
	})

	app.Cron().MustAdd("process-scheduled-sms", "*/1 * * * *", func() {
		if err := services.ProcessDueSchedules(app); err != nil {
			app.Logger().Error("process scheduled SMS", slog.Any("error", err))
		}
	})

	// ── Start ────────────────────────────────────────────────────────
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
