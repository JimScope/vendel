package main

import (
	"ender/handlers"
	"ender/middleware"
	"ender/services"
	"log"
	"log/slog"
	"os"
	"strconv"

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
		services.InitFCM(se.App)

		// Seed first superuser from env vars
		seedSuperuser(se.App)

		// Configure OAuth2 providers from env vars
		configureOAuth(se.App)

		// Configure SMTP from env vars
		configureSMTP(se.App)

		// Configure rate limits
		configureRateLimits(se.App)

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

		// Increment device count before persisting so quota stays consistent
		if err := services.IncrementDeviceCount(e.App, userId); err != nil {
			return err
		}

		// Generate secure API key
		e.Record.Set("api_key", services.GenerateSecureKey("dk_", 32))

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

	// Webhooks: encrypt secret_key on create/update
	app.OnRecordCreate("webhook_configs").BindFunc(func(e *core.RecordEvent) error {
		if secret := e.Record.GetString("secret_key"); secret != "" {
			encrypted, err := services.EncryptSecret(secret)
			if err != nil {
				return err
			}
			e.Record.Set("secret_key", encrypted)
		}
		return e.Next()
	})
	app.OnRecordUpdate("webhook_configs").BindFunc(func(e *core.RecordEvent) error {
		if secret := e.Record.GetString("secret_key"); secret != "" {
			encrypted, err := services.EncryptSecret(secret)
			if err != nil {
				return err
			}
			e.Record.Set("secret_key", encrypted)
		}
		return e.Next()
	})

	// API Keys: generate secure key on create
	app.OnRecordCreate("api_keys").BindFunc(func(e *core.RecordEvent) error {
		e.Record.Set("key", services.GenerateSecureKey("ek_", 32))

		// Unhide so the key is returned in the create response (only shown once)
		e.Record.Unhide("key")

		return e.Next()
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

	app.Cron().MustAdd("retry-failed-sms", "0 */6 * * *", func() {
		if err := services.RetryFailedMessages(app); err != nil {
			app.Logger().Error("retry failed SMS", slog.Any("error", err))
		}
	})

	// ── Start ────────────────────────────────────────────────────────
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// configureOAuth registers OAuth2 providers on the users collection from env vars.
func configureOAuth(app core.App) {
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return
	}

	type providerConfig struct {
		name         string
		clientID     string
		clientSecret string
	}

	providers := []providerConfig{
		{"github", os.Getenv("GITHUB_CLIENT_ID"), os.Getenv("GITHUB_CLIENT_SECRET")},
		{"google", os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")},
	}

	changed := false
	for _, p := range providers {
		if p.clientID == "" || p.clientSecret == "" {
			continue
		}

		existing, found := users.OAuth2.GetProviderConfig(p.name)
		if found && existing.ClientId == p.clientID {
			continue
		}

		if found {
			// Update existing provider in slice
			for i, ep := range users.OAuth2.Providers {
				if ep.Name == p.name {
					users.OAuth2.Providers[i].ClientId = p.clientID
					users.OAuth2.Providers[i].ClientSecret = p.clientSecret
					break
				}
			}
		} else {
			// Append new provider
			users.OAuth2.Providers = append(users.OAuth2.Providers, core.OAuth2ProviderConfig{
				Name:         p.name,
				ClientId:     p.clientID,
				ClientSecret: p.clientSecret,
			})
		}
		changed = true
		app.Logger().Info("Configured OAuth2 provider", slog.String("provider", p.name))
	}

	if changed {
		if err := app.Save(users); err != nil {
			app.Logger().Warn("could not save OAuth2 config", slog.Any("error", err))
		}
	}
}

// configureSMTP sets up SMTP mail settings from environment variables.
func configureSMTP(app core.App) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		if os.Getenv("ENVIRONMENT") != "production" {
			// Default to mailcatcher in dev
			host = "localhost"
		} else {
			return
		}
	}

	port := 1025
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	settings := app.Settings()
	settings.SMTP.Host = host
	settings.SMTP.Port = port
	settings.SMTP.Enabled = true

	if user := os.Getenv("SMTP_USERNAME"); user != "" {
		settings.SMTP.Username = user
	}
	if pass := os.Getenv("SMTP_PASSWORD"); pass != "" {
		settings.SMTP.Password = pass
	}

	// Sender defaults
	if settings.Meta.SenderName == "" || settings.Meta.SenderName == "Support" {
		settings.Meta.SenderName = "Ender"
	}
	if settings.Meta.SenderAddress == "" || settings.Meta.SenderAddress == "support@example.com" {
		settings.Meta.SenderAddress = os.Getenv("FIRST_SUPERUSER")
		if settings.Meta.SenderAddress == "" {
			settings.Meta.SenderAddress = "noreply@ender.app"
		}
	}

	if err := app.Save(settings); err != nil {
		app.Logger().Warn("could not save SMTP config", slog.Any("error", err))
	} else {
		app.Logger().Info("Configured SMTP", slog.String("host", host), slog.Int("port", port))
	}
}

// configureRateLimits enables PocketBase's built-in rate limiter with rules
// for device API endpoints and general API access.
func configureRateLimits(app core.App) {
	settings := app.Settings()

	if settings.RateLimits.Enabled {
		return // already configured (e.g. via admin UI)
	}

	settings.RateLimits.Enabled = true
	settings.RateLimits.Rules = []core.RateLimitRule{
		{Label: "POST /api/sms/report", MaxRequests: 60, Duration: 60},
		{Label: "POST /api/sms/incoming", MaxRequests: 60, Duration: 60},
		{Label: "POST /api/sms/fcm-token", MaxRequests: 10, Duration: 60},
		{Label: "POST /api/sms/send", MaxRequests: 30, Duration: 60},
		{Label: "/api/", MaxRequests: 300, Duration: 60},
	}

	if err := app.Save(settings); err != nil {
		app.Logger().Warn("could not save rate limit config", slog.Any("error", err))
	} else {
		app.Logger().Info("Configured rate limits")
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
		app.Logger().Warn("could not find superusers collection", slog.Any("error", err))
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
		app.Logger().Warn("could not create superuser", slog.Any("error", err))
	} else {
		app.Logger().Info("Created superuser", slog.String("email", email))
	}
}
