package services

import (
	"log/slog"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

const maxRetries = 3

// retryBackoffs defines the minimum wait time before each retry attempt.
var retryBackoffs = []time.Duration{
	15 * time.Minute, // after 1st failure
	1 * time.Hour,    // after 2nd failure
	6 * time.Hour,    // after 3rd failure
}

// isPermanentFailure returns true for errors that should not be retried.
func isPermanentFailure(errMsg string) bool {
	permanent := []string{
		"invalid number",
		"blocked",
		"unsubscribed",
		"blacklisted",
		"not a valid phone",
	}
	lower := strings.ToLower(errMsg)
	for _, p := range permanent {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// RetryFailedMessages retries failed outgoing messages with exponential backoff
// and a maximum of 3 retry attempts. Permanent failures are skipped.
func RetryFailedMessages(app core.App) error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)

	records, err := app.FindRecordsByFilter(
		"sms_messages",
		"status = 'failed' && message_type = 'outgoing' && retry_count < {:maxRetries} && created >= {:cutoff}",
		"", 50, 0,
		dbx.Params{"maxRetries": maxRetries, "cutoff": cutoff},
	)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	retried := 0
	skipped := 0
	for _, record := range records {
		// Skip permanent failures
		if isPermanentFailure(record.GetString("error_message")) {
			skipped++
			continue
		}

		// Enforce exponential backoff based on retry_count
		retryCount := record.GetInt("retry_count")
		if retryCount > 0 && retryCount <= len(retryBackoffs) {
			lastRetry := record.GetDateTime("last_retry_at").Time()
			if !lastRetry.IsZero() {
				requiredWait := retryBackoffs[retryCount-1]
				if now.Sub(lastRetry) < requiredWait {
					continue // not enough time has passed
				}
			}
		}

		record.Set("status", "pending")
		record.Set("retry_count", retryCount+1)
		record.Set("last_retry_at", types.NowDateTime())
		record.Set("error_message", "")
		if err := app.Save(record); err == nil {
			retried++
		}
	}

	app.Logger().Info("Retried failed SMS messages",
		slog.Int("retried", retried), slog.Int("skipped_permanent", skipped))
	return nil
}
