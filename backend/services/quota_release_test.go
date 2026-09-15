package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"vendel/services"
	"vendel/services/smsprovider"
)

// errProvider is a stub Provider whose Send always fails, exercising the REL-2
// transport-error path in dispatchOne (reached via DispatchProviderMessages).
type errProvider struct{}

func (errProvider) Name() string       { return "err-provider" }
func (errProvider) IsConfigured() bool { return true }
func (errProvider) Send(ctx context.Context, req smsprovider.SendRequest) (*smsprovider.SendResult, error) {
	return nil, errors.New("boom: network unreachable")
}

// REL-1(a): a never-sent outgoing message that reaches a terminal "failed"
// state refunds exactly one reserved credit (models pending→failed and
// sending→failed rescues, both of which route through MarkMessageTerminal).
func TestMarkMessageTerminal_RefundsUnsentTerminalFailure(t *testing.T) {
	app := newTestApp(t)
	userId := makeTestUser(t, app, "refund@test.com")

	if err := services.ReserveSMSQuota(app, userId, 1); err != nil {
		t.Fatalf("ReserveSMSQuota: %v", err)
	}
	if got := quotaUsed(t, app, userId); got != 1 {
		t.Fatalf("after reserve: used = %d, want 1", got)
	}

	// retry_count >= SMSMaxRetries makes the failure terminal (never resurrected
	// by RetryFailedMessages). status "pending" + no sent_at models a message
	// that never left the gateway.
	msg := makeOutgoingMessage(t, app, userId, "pending", services.SMSMaxRetries, false)
	if err := services.MarkMessageTerminal(app, msg, "failed", "no device available"); err != nil {
		t.Fatalf("MarkMessageTerminal: %v", err)
	}

	if got := quotaUsed(t, app, userId); got != 0 {
		t.Fatalf("after terminal failure: used = %d, want 0 (credit refunded)", got)
	}
}

// REL-1(b): a message that actually reached "sent" (sent_at set) and later
// fails delivery keeps its already-spent credit — no refund.
func TestMarkMessageTerminal_SentThenFailedDoesNotRefund(t *testing.T) {
	app := newTestApp(t)
	userId := makeTestUser(t, app, "nosentrefund@test.com")

	if err := services.ReserveSMSQuota(app, userId, 1); err != nil {
		t.Fatalf("ReserveSMSQuota: %v", err)
	}

	// status "sent" with sent_at set, terminal retry_count: this is the AEUM
	// delivery-failure shape (sent successfully, provider later reports failure).
	msg := makeOutgoingMessage(t, app, userId, "sent", services.SMSMaxRetries, true)
	if err := services.MarkMessageTerminal(app, msg, "failed", "provider delivery failure: TEXT_BLOCKED"); err != nil {
		t.Fatalf("MarkMessageTerminal: %v", err)
	}

	if got := quotaUsed(t, app, userId); got != 1 {
		t.Fatalf("after sent→failed: used = %d, want 1 (no refund)", got)
	}
}

// REL-1(c): a retryable failure (retry_count below the max, recent) must NOT
// refund early — RetryFailedMessages will resurrect it, and an early refund
// would under-count a message that later succeeds.
func TestMarkMessageTerminal_RetryableFailureDoesNotRefund(t *testing.T) {
	app := newTestApp(t)
	userId := makeTestUser(t, app, "retryable@test.com")

	if err := services.ReserveSMSQuota(app, userId, 1); err != nil {
		t.Fatalf("ReserveSMSQuota: %v", err)
	}

	msg := makeOutgoingMessage(t, app, userId, "assigned", 0, false)
	if err := services.MarkMessageTerminal(app, msg, "failed", "temporary send error"); err != nil {
		t.Fatalf("MarkMessageTerminal: %v", err)
	}

	if got := quotaUsed(t, app, userId); got != 1 {
		t.Fatalf("after retryable failure: used = %d, want 1 (no early refund)", got)
	}
}

// REL-2: a provider.Send transport error marks the message terminal "failed"
// (instead of orphaning it in "assigned") and, when the message is already
// past its retries, releases the reserved credit via REL-1.
func TestDispatchProviderMessages_SendErrorMarksFailed(t *testing.T) {
	app := newTestApp(t)
	userId := makeTestUser(t, app, "dispatch@test.com")

	if err := services.ReserveSMSQuota(app, userId, 1); err != nil {
		t.Fatalf("ReserveSMSQuota: %v", err)
	}

	// retry_count >= SMSMaxRetries so the send error is terminal and REL-1 fires.
	msg := makeOutgoingMessage(t, app, userId, "assigned", services.SMSMaxRetries, false)

	services.DispatchProviderMessages(app, errProvider{}, []*core.Record{msg})

	reloaded, err := app.FindRecordById("sms_messages", msg.Id)
	if err != nil {
		t.Fatalf("reload message: %v", err)
	}
	if got := reloaded.GetString("status"); got != "failed" {
		t.Fatalf("status = %q, want %q", got, "failed")
	}
	if got := reloaded.GetString("error_message"); got == "" {
		t.Fatalf("error_message is empty, want provider send failure detail")
	}
	if got := quotaUsed(t, app, userId); got != 0 {
		t.Fatalf("after terminal provider error: used = %d, want 0 (credit refunded)", got)
	}
}
