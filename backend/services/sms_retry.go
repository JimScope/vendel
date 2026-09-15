package services

import (
	"log/slog"
	"strings"
	"time"
	"vendel/services/smsprovider"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// isPermanentFailure returns true for errors that should not be retried.
// "provider delivery failure" covers terminal delivery events reported by an
// external provider (e.g. AEUM via SNS): the provider already exhausted its
// own retries, so re-dispatching would only bill another SMS.
func isPermanentFailure(errMsg string) bool {
	permanent := []string{
		"invalid number",
		"blocked",
		"unsubscribed",
		"blacklisted",
		"not a valid phone",
		"provider delivery failure",
	}
	lower := strings.ToLower(errMsg)
	for _, p := range permanent {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// isTerminalFailure reports whether a message that just reached "failed" is
// genuinely terminal — i.e. RetryFailedMessages will never resurrect it. It is
// the exact inverse of the retry-eligibility filter used by RetryFailedMessages
// (retry_count < max && created >= cutoff && !permanent), so callers can release
// the up-front quota reservation exactly once, at the moment the message stops
// being retryable.
func isTerminalFailure(msg *core.Record, errMsg string) bool {
	if isPermanentFailure(errMsg) {
		return true
	}
	if msg.GetInt("retry_count") >= SMSMaxRetries {
		return true
	}
	cutoff := FilterTime(time.Now().UTC().Add(-SMSRetryCutoff))
	return msg.GetString("created") < cutoff
}

// RetryFailedMessages retries failed outgoing messages with exponential backoff
// and a maximum of SMSMaxRetries attempts. Permanent failures are skipped.
// Retried messages go back to "assigned" and are re-dispatched (FCM tickle /
// provider send); modem agents are notified by the realtime hook on save.
// It also rescues "pending" messages created when the user had no device,
// assigning them as soon as a device becomes available.
func RetryFailedMessages(app core.App) error {
	cutoff := FilterTime(time.Now().UTC().Add(-SMSRetryCutoff))
	now := time.Now().UTC()

	// PERF-1: cache device resolution per userId for this cron pass so a backlog
	// of N failed messages for the same user costs one resolveDevices query, not
	// N. The cache is created fresh each pass and shared with the pending rescue,
	// so it never serves stale device sets across passes.
	deviceCache := make(map[string][]*core.Record)

	retried := 0
	skipped := 0
	requeued := make([]*core.Record, 0)

	// REL-4: drain the whole eligible backlog in bounded batches instead of a
	// single 50-row query that silently ignored anything beyond it.
	err := processInBatches(app, "sms_messages",
		"status = 'failed' && message_type = 'outgoing' && retry_count < {:maxRetries} && created >= {:cutoff}",
		"created",
		dbx.Params{"maxRetries": SMSMaxRetries, "cutoff": cutoff},
		func(record *core.Record) bool {
			// Skip permanent failures — they stay 'failed'.
			if isPermanentFailure(record.GetString("error_message")) {
				skipped++
				return false
			}

			// Enforce exponential backoff based on retry_count.
			retryCount := record.GetInt("retry_count")
			if retryCount > 0 && retryCount <= len(SMSRetryBackoffs) {
				lastRetry := record.GetDateTime("last_retry_at").Time()
				if !lastRetry.IsZero() {
					requiredWait := SMSRetryBackoffs[retryCount-1]
					if now.Sub(lastRetry) < requiredWait {
						return false // not enough time has passed; stays 'failed'
					}
				}
			}

			// Messages can lose their device (e.g. it was deleted) — re-resolve
			// so the retry has a transport to go through.
			if record.GetString("device") == "" {
				if !assignAvailableDevice(app, record, deviceCache) {
					return false // no device; stays 'failed'
				}
			}

			record.Set("status", "assigned")
			record.Set("retry_count", retryCount+1)
			record.Set("last_retry_at", types.NowDateTime())
			record.Set("error_message", "")
			if err := app.Save(record); err != nil {
				return false // save failed; stays 'failed'
			}
			retried++
			requeued = append(requeued, record)
			return true // left the 'failed' set
		})
	if err != nil {
		return err
	}

	requeued = append(requeued, rescuePendingMessages(app, cutoff, deviceCache)...)
	requeued = append(requeued, rescueStaleSendingMessages(app)...)

	// Re-dispatch outside the loop: one FCM tickle per device, one provider
	// batch per type, instead of per-message traffic.
	if len(requeued) > 0 {
		dispatchOutgoing(app, requeued)
	}

	app.Logger().Info("Retried failed SMS messages",
		slog.Int("retried", retried), slog.Int("skipped_permanent", skipped))
	return nil
}

// rescuePendingMessages assigns devices to outgoing messages stuck in
// "pending" (created while the user had no usable device). Messages older
// than the retry cutoff are marked failed so their state is honest and the
// sms_failed webhook fires.
func rescuePendingMessages(app core.App, cutoff string, deviceCache map[string][]*core.Record) []*core.Record {
	rescued := make([]*core.Record, 0)

	// REL-4: drain all pending messages in bounded batches, not just the first 50.
	err := processInBatches(app, "sms_messages",
		"status = 'pending' && message_type = 'outgoing'",
		"created", nil,
		func(record *core.Record) bool {
			if assignAvailableDevice(app, record, deviceCache) {
				record.Set("status", "assigned")
				if err := app.Save(record); err != nil {
					return false // stays 'pending'
				}
				rescued = append(rescued, record)
				return true // left the 'pending' set
			}

			// Still no device — fail messages past the retry window instead of
			// leaving them in limbo forever.
			if record.GetString("created") < cutoff {
				if err := MarkMessageTerminal(app, record, "failed", "no device available"); err != nil {
					app.Logger().Warn("failed to expire pending message",
						slog.String("message", record.Id), slog.Any("error", err))
					return false // still 'pending'
				}
				return true // left the 'pending' set (now 'failed')
			}
			return false // within window, no device yet — stays 'pending'
		})
	if err != nil {
		app.Logger().Warn("failed to query pending messages", slog.Any("error", err))
	}
	return rescued
}

// rescueStaleSendingMessages re-queues outgoing messages stuck in "sending".
// Agents move assigned → sending when they claim a batch via
// /api/sms/pending; if the agent dies before sending (or before reporting),
// nothing else ever touches the record — pending fetches only return
// "assigned" — so without this rescue the message is stranded forever.
// Re-queueing after the stale window accepts a small duplicate-SMS risk
// (sent but never reported) in exchange for never losing a message.
func rescueStaleSendingMessages(app core.App) []*core.Record {
	staleCutoff := FilterTime(time.Now().UTC().Add(-SMSSendingStaleAfter))
	rescued := make([]*core.Record, 0)

	// REL-4: drain all stale-sending messages in bounded batches. Every record
	// here is acted on (re-queued or expired), so its `updated` timestamp moves
	// past the stale cutoff and it leaves the eligible set.
	err := processInBatches(app, "sms_messages",
		"status = 'sending' && message_type = 'outgoing' && updated < {:staleCutoff}",
		"created", dbx.Params{"staleCutoff": staleCutoff},
		func(record *core.Record) bool {
			retryCount := record.GetInt("retry_count")
			if retryCount >= SMSMaxRetries {
				if err := MarkMessageTerminal(app, record, "failed", "agent claimed the message but never reported a result"); err != nil {
					app.Logger().Warn("failed to expire stale sending message",
						slog.String("message", record.Id), slog.Any("error", err))
					return false // still 'sending'
				}
				return true // left the 'sending' set (now 'failed')
			}

			record.Set("status", "assigned")
			record.Set("retry_count", retryCount+1)
			record.Set("last_retry_at", types.NowDateTime())
			if err := app.Save(record); err != nil {
				return false // still 'sending'
			}
			rescued = append(rescued, record)
			return true // left the 'sending' set
		})
	if err != nil {
		app.Logger().Warn("failed to query stale sending messages", slog.Any("error", err))
	}
	if len(rescued) > 0 {
		app.Logger().Info("rescued stale sending messages", slog.Int("count", len(rescued)))
	}
	return rescued
}

// assignAvailableDevice resolves a device for the message's user and sets the
// device + from_number fields. Returns false when no device is available.
//
// PERF-1: device resolution is cached per userId in deviceCache for the current
// cron pass, so re-resolving the same per-user device set for every record is
// avoided. A nil/empty result is cached too (comma-ok distinguishes "resolved
// to none" from "not yet resolved"), so users with no device cost one query,
// not one per record.
func assignAvailableDevice(app core.App, record *core.Record, deviceCache map[string][]*core.Record) bool {
	userId := record.GetString("user")
	devices, ok := deviceCache[userId]
	if !ok {
		resolved, err := resolveDevices(app, userId, "", smsprovider.DefaultAEUM())
		if err != nil {
			resolved = nil
		}
		devices = resolved
		deviceCache[userId] = devices
	}
	if len(devices) == 0 {
		return false
	}
	device := devices[0]
	record.Set("device", device.Id)
	record.Set("from_number", device.GetString("phone_number"))
	return true
}
