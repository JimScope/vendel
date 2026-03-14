package handlers

import (
	"vendel/middleware"
	"vendel/services"
	"net/http"
	"regexp"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// RegisterSMSRoutes registers custom SMS API routes.
func RegisterSMSRoutes(se *core.ServeEvent) {
	// POST /api/sms/send — Send SMS (auth: JWT or API key)
	se.Router.POST("/api/sms/send", func(e *core.RequestEvent) error {
		userId, err := middleware.ResolveAuthOrAPIKey(e)
		if err != nil {
			return apis.NewUnauthorizedError("Authentication required", nil)
		}

		var body struct {
			Recipients []string `json:"recipients"`
			Body       string   `json:"body"`
			DeviceID   string   `json:"device_id"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		if len(body.Recipients) == 0 {
			return apis.NewBadRequestError("At least one recipient required", nil)
		}
		for _, r := range body.Recipients {
			if !e164Regex.MatchString(r) {
				return apis.NewBadRequestError("Invalid phone number: "+r+". Must be E.164 format (e.g. +1234567890)", nil)
			}
		}
		if body.Body == "" {
			return apis.NewBadRequestError("Message body required", nil)
		}
		if len(body.Body) > 1600 {
			return apis.NewBadRequestError("Message body exceeds 1600 character limit", nil)
		}

		messages, err := services.SendSMS(e.App, userId, body.Recipients, body.Body, body.DeviceID)
		if err != nil {
			return handleServiceError(e, err)
		}

		// Build response
		ids := make([]string, len(messages))
		for i, m := range messages {
			ids[i] = m.Id
		}

		var batchId string
		if len(messages) > 0 {
			batchId = messages[0].GetString("batch_id")
		}

		return e.JSON(http.StatusOK, map[string]any{
			"batch_id":         batchId,
			"message_ids":      ids,
			"recipients_count": len(body.Recipients),
			"status":           "accepted",
		})
	})

	// POST /api/sms/report — Device ACK callback (auth: device API key)
	se.Router.POST("/api/sms/report", func(e *core.RequestEvent) error {
		device, err := middleware.AuthenticateDevice(e)
		if err != nil {
			return apis.NewUnauthorizedError("Invalid API key", nil)
		}

		var body struct {
			MessageID    string `json:"message_id"`
			Status       string `json:"status"`
			ErrorMessage string `json:"error_message"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		// Only allow terminal statuses from device reports
		validStatuses := map[string]bool{"sent": true, "delivered": true, "failed": true}
		if !validStatuses[body.Status] {
			return apis.NewBadRequestError("Invalid status: must be sent, delivered, or failed", nil)
		}

		if err := services.ProcessSMSAck(e.App, device.Id, body.MessageID, body.Status, body.ErrorMessage); err != nil {
			return apis.NewNotFoundError(err.Error(), nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"success":    true,
			"message_id": body.MessageID,
			"status":     body.Status,
		})
	})

	// POST /api/sms/incoming — Incoming SMS from device (auth: device API key)
	se.Router.POST("/api/sms/incoming", func(e *core.RequestEvent) error {
		device, err := middleware.AuthenticateDevice(e)
		if err != nil {
			return apis.NewUnauthorizedError("Invalid API key", nil)
		}

		var body struct {
			FromNumber string `json:"from_number"`
			Body       string `json:"body"`
			Timestamp  string `json:"timestamp"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		userId := device.GetString("user")
		message, err := services.HandleIncomingSMS(e.App, userId, device.Id, body.FromNumber, body.Body, body.Timestamp)
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, err.Error(), nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"success":    true,
			"message_id": message.Id,
		})
	})

	// GET /api/sms/pending — Pending messages for devices: Android (post-FCM tickle) and modems (startup recovery) (auth: device API key)
	se.Router.GET("/api/sms/pending", func(e *core.RequestEvent) error {
		device, err := middleware.AuthenticateDevice(e)
		if err != nil {
			return apis.NewUnauthorizedError("Invalid API key", nil)
		}

		claimed, err := services.ClaimPendingMessages(e.App, device.Id)
		if err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to claim messages", nil)
		}

		type pendingMsg struct {
			MessageID string `json:"message_id"`
			Recipient string `json:"recipient"`
			Body      string `json:"body"`
		}
		msgs := make([]pendingMsg, 0, len(claimed))
		for _, r := range claimed {
			msgs = append(msgs, pendingMsg{
				MessageID: r.Id,
				Recipient: r.GetString("to"),
				Body:      services.GetRecordBody(r),
			})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"device_id": device.Id,
			"messages":  msgs,
		})
	})

	// POST /api/sms/fcm-token — Update device FCM token (auth: device API key)
	se.Router.POST("/api/sms/fcm-token", func(e *core.RequestEvent) error {
		device, err := middleware.AuthenticateDevice(e)
		if err != nil {
			return apis.NewUnauthorizedError("Invalid API key", nil)
		}

		var body struct {
			FCMToken string `json:"fcm_token"`
		}
		if err := e.BindBody(&body); err != nil {
			return apis.NewBadRequestError("Invalid request body", nil)
		}

		device.Set("fcm_token", body.FCMToken)
		if err := e.App.Save(device); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to update token", nil)
		}

		return e.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}
