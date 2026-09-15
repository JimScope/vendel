package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"
)

// ProcessDueSchedules finds and dispatches all scheduled SMS that are due.
func ProcessDueSchedules(app core.App) error {
	now := FilterNow()

	dispatched := 0
	failed := 0

	// REL-4: drain all due schedules in bounded batches instead of a single
	// 50-row query that silently left the rest for the next run.
	err := processInBatches(app, "scheduled_sms",
		"status = 'active' && next_run_at != '' && next_run_at <= {:now}",
		"next_run_at",
		dbx.Params{"now": now},
		func(record *core.Record) bool {
			// Decode recipients
			var recipients []string
			recipientsJSON := record.GetString("recipients")
			if err := json.Unmarshal([]byte(recipientsJSON), &recipients); err != nil {
				app.Logger().Warn("scheduled SMS: invalid recipients JSON",
					slog.String("id", record.Id), slog.Any("error", err))
				failed++
				return false // unchanged; skip past it this pass
			}

			userId := record.GetString("user")
			body := GetRecordBody(record)
			deviceId := record.GetString("device_id")

			// Send the SMS
			if _, err := SendSMS(app, userId, recipients, body, deviceId, nil); err != nil {
				app.Logger().Warn("scheduled SMS: send failed",
					slog.String("id", record.Id), slog.Any("error", err))
				failed++
				return false // next_run_at unchanged; skip past it this pass
			}

			// Update record after successful send
			record.Set("last_run_at", time.Now().UTC().Format(time.RFC3339))

			// leftSet tracks whether the record still matches the due filter after
			// the update. A recurring schedule whose next run can't be computed
			// keeps next_run_at <= now, so it stays eligible — return false there
			// so it is not re-sent again within the same pass (send once per pass).
			leftSet := true
			scheduleType := record.GetString("schedule_type")
			if scheduleType == "one_time" {
				record.Set("status", "completed")
				record.Set("next_run_at", "")
			} else if scheduleType == "recurring" {
				cronExpr := record.GetString("cron_expression")
				tz := record.GetString("timezone")
				if tz == "" {
					tz = "UTC"
				}
				nextRun, err := ComputeNextRun(cronExpr, tz)
				if err != nil {
					app.Logger().Warn("scheduled SMS: failed to compute next run",
						slog.String("id", record.Id), slog.Any("error", err))
					leftSet = false
				} else {
					record.Set("next_run_at", nextRun)
				}
			}

			if err := app.Save(record); err != nil {
				app.Logger().Warn("scheduled SMS: failed to update record",
					slog.String("id", record.Id), slog.Any("error", err))
				failed++
				return false
			}
			dispatched++
			return leftSet
		})
	if err != nil {
		return err
	}

	if dispatched > 0 || failed > 0 {
		app.Logger().Info("Processed scheduled SMS",
			slog.Int("dispatched", dispatched), slog.Int("failed", failed))
	}

	return nil
}

// ComputeNextRun calculates the next occurrence of a cron expression in the
// given IANA timezone. Returns a UTC RFC3339 string.
func ComputeNextRun(cronExpr, timezone string) (string, error) {
	schedule, err := cron.NewSchedule(cronExpr)
	if err != nil {
		return "", fmt.Errorf("invalid cron expression: %w", err)
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone: %w", err)
	}

	// Start from the next minute
	now := time.Now().In(loc)
	candidate := now.Truncate(time.Minute).Add(time.Minute)

	// Cap at 366 days to avoid infinite loops
	limit := candidate.Add(366 * 24 * time.Hour)
	for candidate.Before(limit) {
		if schedule.IsDue(cron.NewMoment(candidate)) {
			return candidate.UTC().Format(time.RFC3339), nil
		}
		candidate = candidate.Add(time.Minute)
	}

	return "", fmt.Errorf("no matching time found within 366 days")
}

// ValidateCronExpression checks if a cron expression is valid.
func ValidateCronExpression(cronExpr string) error {
	_, err := cron.NewSchedule(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}
	return nil
}
