package main

import (
	"vendel/services"
	"log/slog"
	"os"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

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

	// Sender defaults — use app_name from system_config
	appName := services.GetSystemConfigValue(app, "app_name")
	if appName == "" {
		appName = "Vendel"
	}
	settings.Meta.AppName = appName
	if settings.Meta.SenderName == "" || settings.Meta.SenderName == "Support" || settings.Meta.SenderName == "Acme" {
		settings.Meta.SenderName = appName
	}
	if settings.Meta.SenderAddress == "" || settings.Meta.SenderAddress == "support@example.com" {
		settings.Meta.SenderAddress = os.Getenv("FIRST_SUPERUSER")
		if settings.Meta.SenderAddress == "" {
			settings.Meta.SenderAddress = "noreply@vendel.cc"
		}
	}

	if err := app.Save(settings); err != nil {
		app.Logger().Warn("could not save SMTP config", slog.Any("error", err))
	} else {
		app.Logger().Info("Configured SMTP", slog.String("host", host), slog.Int("port", port))
	}
}

// configureAuthSecurity hardens password policy and token lifetime for the users collection.
func configureAuthSecurity(app core.App) {
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return
	}

	changed := false

	// Enforce minimum 10-character passwords
	if pwField, ok := users.Fields.GetByName("password").(*core.PasswordField); ok {
		if pwField.Min < 10 {
			pwField.Min = 10
			changed = true
		}
	}

	// Reduce auth token lifetime from 7 days to 24 hours
	if users.AuthToken.Duration > 86400 {
		users.AuthToken.Duration = 86400
		changed = true
	}

	if changed {
		if err := app.Save(users); err != nil {
			app.Logger().Warn("could not save auth security config", slog.Any("error", err))
		} else {
			app.Logger().Info("Configured auth security (min password: 10, token: 24h)")
		}
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
		{Label: "POST /api/collections/users/auth-with-password", MaxRequests: 5, Duration: 300},
		{Label: "POST /api/sms/report", MaxRequests: 60, Duration: 60},
		{Label: "POST /api/sms/incoming", MaxRequests: 60, Duration: 60},
		{Label: "POST /api/sms/fcm-token", MaxRequests: 10, Duration: 60},
		{Label: "POST /api/sms/send", MaxRequests: 30, Duration: 60},
		{Label: "GET /api/sms/pending", MaxRequests: 30, Duration: 60},
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
