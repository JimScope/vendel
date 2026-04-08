package cronjobs

import (
	"log/slog"
	"vendel/services"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterCronJobs registers all periodic background tasks.
func RegisterCronJobs(app *pocketbase.PocketBase) {
	register(app, "daily-quota-reset", "0 0 * * *", services.ResetMonthlyQuotas, "quota reset failed")
	register(app, "daily-renewal-check", "0 8 * * *", services.CheckRenewals, "renewal check failed")
	register(app, "retry-failed-sms", "*/15 * * * *", services.RetryFailedMessages, "retry failed SMS")
	register(app, "retry-failed-webhooks", "*/1 * * * *", services.RetryFailedWebhooks, "retry failed webhooks")
	register(app, "process-scheduled-sms", "*/1 * * * *", services.ProcessDueSchedules, "process scheduled SMS")
	register(app, "purge-expired-data", "0 3 * * *", services.PurgeExpiredData, "purge expired data failed")
}

// register is a helper that wraps a service function in standard cron
// boilerplate: call the function, log on error.
func register(app *pocketbase.PocketBase, name, schedule string, fn func(core.App) error, errMsg string) {
	app.Cron().MustAdd(name, schedule, func() {
		if err := fn(app); err != nil {
			app.Logger().Error(errMsg, slog.Any("error", err))
		}
	})
}
