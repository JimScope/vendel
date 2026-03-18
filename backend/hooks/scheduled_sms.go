package hooks

import (
	"fmt"
	"vendel/services"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterScheduledSMSHooks computes next_run_at on create and update
// for scheduled_sms records.
func RegisterScheduledSMSHooks(app *pocketbase.PocketBase) {
	app.OnRecordCreate("scheduled_sms").BindFunc(func(e *core.RecordEvent) error {
		userId := e.Record.GetString("user")
		if err := services.CheckScheduledSMSQuota(e.App, userId); err != nil {
			return err
		}
		if e.Record.GetString("timezone") == "" {
			e.Record.Set("timezone", "UTC")
		}
		if e.Record.GetString("status") == "" {
			e.Record.Set("status", "active")
		}
		if err := computeNextRunAt(e.Record); err != nil {
			return err
		}
		return e.Next()
	})

	app.OnRecordUpdate("scheduled_sms").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("timezone") == "" {
			e.Record.Set("timezone", "UTC")
		}
		if err := computeNextRunAt(e.Record); err != nil {
			return err
		}
		return e.Next()
	})
}

// computeNextRunAt sets next_run_at based on schedule_type.
// For one_time schedules it copies scheduled_at; for recurring it
// evaluates the cron expression in the record's timezone.
func computeNextRunAt(record *core.Record) error {
	scheduleType := record.GetString("schedule_type")
	if scheduleType == "one_time" {
		record.Set("next_run_at", record.GetString("scheduled_at"))
	} else if scheduleType == "recurring" {
		cronExpr := record.GetString("cron_expression")
		tz := record.GetString("timezone")
		nextRun, err := services.ComputeNextRun(cronExpr, tz)
		if err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		record.Set("next_run_at", nextRun)
	}
	return nil
}
