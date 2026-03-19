export const WEBHOOK_EVENTS = [
  "sms_received",
  "sms_sent",
  "sms_delivered",
  "sms_failed",
] as const

export type WebhookEvent = (typeof WEBHOOK_EVENTS)[number]

export const WEBHOOK_EVENT_LABELS: Record<WebhookEvent, string> = {
  sms_received: "SMS Received",
  sms_sent: "SMS Sent",
  sms_delivered: "SMS Delivered",
  sms_failed: "SMS Failed",
}

export const WEBHOOK_EVENT_DESCRIPTIONS: Record<WebhookEvent, string> = {
  sms_received: "Triggered when an incoming SMS arrives on a device",
  sms_sent: "Triggered when an outgoing SMS is confirmed as sent",
  sms_delivered: "Triggered when an outgoing SMS is confirmed as delivered",
  sms_failed: "Triggered when an SMS fails to send or deliver",
}

export const WEBHOOK_EVENT_PAYLOADS: Record<WebhookEvent, object> = {
  sms_received: {
    event: "sms_received",
    message_id: "abc123",
    timestamp: "2026-01-15T10:30:00Z",
    from: "+1234567890",
    body: "Message content (if include_body is enabled)",
  },
  sms_sent: {
    event: "sms_sent",
    message_id: "xyz789",
    timestamp: "2026-01-15T10:30:00Z",
    to: "+1234567890",
    status: "sent",
    sent_at: "2026-01-15T10:30:05Z",
    body: "Message content (if include_body is enabled)",
  },
  sms_delivered: {
    event: "sms_delivered",
    message_id: "xyz789",
    timestamp: "2026-01-15T10:30:00Z",
    to: "+1234567890",
    status: "delivered",
    delivered_at: "2026-01-15T10:30:10Z",
    body: "Message content (if include_body is enabled)",
  },
  sms_failed: {
    event: "sms_failed",
    message_id: "xyz789",
    timestamp: "2026-01-15T10:30:00Z",
    to: "+1234567890",
    status: "failed",
    error_message: "Device offline",
    body: "Message content (if include_body is enabled)",
  },
}
