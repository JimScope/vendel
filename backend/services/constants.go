package services

import "time"

// ── Application ─────────────────────────────────────────────────────

const DefaultAppName = "Vendel"

// ── SMS ─────────────────────────────────────────────────────────────

const MaxMessageBodyLength = 1600

// MaxRecipientsPerRequest caps how many recipients a single send request may
// target (after group expansion and de-duplication). Without it a request —
// or a group that expands to thousands of contacts — could reserve a huge
// quota block and create thousands of rows in one call. The cap is enforced
// before any quota is reserved or record created, so an over-sized request is
// rejected cleanly with a 400-style error.
const MaxRecipientsPerRequest = 500

// ── API Keys & Device Keys ──────────────────────────────────────────

const (
	APIKeyPrefix     = "vk_"
	DeviceKeyPrefix  = "dk_"
	GeneratedKeyLen  = 32
	KeyPrefixDisplay = 10 // characters shown in "vk_XXXXXX..."
)

// ── Auth ────────────────────────────────────────────────────────────

const (
	MinPasswordLength      = 10
	AuthTokenDurationSecs  = 86400 // 24 hours
	DefaultAPIKeyExpiryYrs = 1     // years from creation
)

// ── Webhook Delivery ────────────────────────────────────────────────

const (
	WebhookMaxRetries       = 3
	WebhookDialTimeout      = 10 * time.Second
	WebhookDefaultTimeout   = 10 // seconds (per-webhook configurable)
	WebhookMaxTimeout       = 30 // seconds (hard cap)
	WebhookMaxRedirects     = 3
	WebhookResponseMaxBytes = 2048
	WebhookResponseMaxChars = 2000
	WebhookIdleConnTimeout  = 90 * time.Second
	WebhookMaxIdleConns     = 20
	WebhookMaxIdlePerHost   = 5
)

// ── Webhook Retry Backoffs ──────────────────────────────────────────

var WebhookRetryBackoffs = []time.Duration{
	1 * time.Minute,  // after 1st failure
	5 * time.Minute,  // after 2nd failure
	15 * time.Minute, // after 3rd failure
}

// WebhookRetryJitter is the maximum fraction by which each retry backoff is
// randomly perturbed (±20%). Without it, every webhook that failed against the
// same downed host during one burst shares an identical next_retry_at and they
// all retry in a synchronized spike the moment the host recovers — the classic
// thundering herd. Jittering spreads those retries across a window so a
// recovering host is not hammered. The floor factor (1-jitter) stays positive,
// so the backoff never collapses to zero and keeps growing across attempts.
const WebhookRetryJitter = 0.2

// ── SMS Retry ───────────────────────────────────────────────────────

const (
	SMSMaxRetries  = 3
	SMSRetryCutoff = 24 * time.Hour // only retry messages younger than this

	// SMSSendingStaleAfter marks "sending" messages as stuck: an agent claims
	// them via /api/sms/pending and a healthy send completes in well under a
	// minute, so anything still "sending" after this window belongs to an
	// agent that died mid-batch and must be re-queued.
	SMSSendingStaleAfter = 10 * time.Minute

	// SMSClaimBatchSize caps how many "assigned" messages an agent flips to
	// "sending" in one /api/sms/pending call. Kept small enough that even a
	// slow modem (multipart / rate-limited link) drains a full claim well
	// within SMSSendingStaleAfter, so rescueStaleSendingMessages never
	// re-queues a batch a healthy-but-slow agent is still working through —
	// which would duplicate sends. The agent simply claims again next poll.
	SMSClaimBatchSize = 20
)

var SMSRetryBackoffs = []time.Duration{
	15 * time.Minute, // after 1st failure
	1 * time.Hour,    // after 2nd failure
	6 * time.Hour,    // after 3rd failure
}

// ── Cron Batch Draining ─────────────────────────────────────────────

const (
	// CronDrainBatchSize is how many records a batched cron drain fetches per
	// query. CronMaxDrainBatches caps the number of batches processed in a
	// single pass, so a large backlog is drained (up to
	// CronDrainBatchSize*CronMaxDrainBatches records) without an unbounded loop.
	CronDrainBatchSize  = 50
	CronMaxDrainBatches = 10 // up to 500 records per drain per pass
)

// ── External Service Timeouts ───────────────────────────────────────

const (
	PaymentClientTimeout = 30 * time.Second
	FCMContextTimeout    = 10 * time.Second
)
