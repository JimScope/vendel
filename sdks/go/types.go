package vendel

// SendSMSRequest is the payload for sending an SMS.
type SendSMSRequest struct {
	Recipients []string `json:"recipients"`
	Body       string   `json:"body"`
	DeviceID   string   `json:"device_id,omitempty"`
}

// SendSMSResponse is returned after a successful send.
type SendSMSResponse struct {
	BatchID         string   `json:"batch_id"`
	MessageIDs      []string `json:"message_ids"`
	RecipientsCount int      `json:"recipients_count"`
	Status          string   `json:"status"`
}

// Quota represents the user's current plan limits and usage.
type Quota struct {
	Plan             string `json:"plan"`
	SMSSentThisMonth int    `json:"sms_sent_this_month"`
	MaxSMSPerMonth   int    `json:"max_sms_per_month"`
	DevicesRegistered int   `json:"devices_registered"`
	MaxDevices       int    `json:"max_devices"`
	ResetDate        string `json:"reset_date"`
}
