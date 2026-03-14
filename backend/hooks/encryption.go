package hooks

import (
	"fmt"
	"vendel/services"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterEncryptionHooks encrypts the body field on save and decrypts
// it on API responses for sms_messages, sms_templates, and scheduled_sms.
func RegisterEncryptionHooks(app *pocketbase.PocketBase) {
	// Encrypt body on save for templates and scheduled SMS
	encryptBodyOnSave := func(e *core.RecordEvent) error {
		body := e.Record.GetString("body")
		if body != "" && !services.IsBodyEncrypted(body) {
			encrypted, err := services.EncryptBody(body)
			if err != nil {
				return fmt.Errorf("body encryption failed: %w", err)
			}
			e.Record.Set("body", encrypted)
		}
		return e.Next()
	}
	app.OnRecordCreate("sms_templates").BindFunc(encryptBodyOnSave)
	app.OnRecordUpdate("sms_templates").BindFunc(encryptBodyOnSave)
	app.OnRecordCreate("scheduled_sms").BindFunc(encryptBodyOnSave)
	app.OnRecordUpdate("scheduled_sms").BindFunc(encryptBodyOnSave)

	// Encrypt body + compute blind index for sms_messages
	encryptBodyWithHash := func(e *core.RecordEvent) error {
		body := e.Record.GetString("body")
		if body != "" && !services.IsBodyEncrypted(body) {
			hash, err := services.ComputeBodyHash(body)
			if err != nil {
				return fmt.Errorf("body hash failed: %w", err)
			}
			encrypted, err := services.EncryptBody(body)
			if err != nil {
				return fmt.Errorf("body encryption failed: %w", err)
			}
			e.Record.Set("body_hash", hash)
			e.Record.Set("body", encrypted)
		}
		return e.Next()
	}
	app.OnRecordCreate("sms_messages").BindFunc(encryptBodyWithHash)
	app.OnRecordUpdate("sms_messages").BindFunc(encryptBodyWithHash)

	// Decrypt body for API responses (frontend always sees plaintext)
	app.OnRecordEnrich("sms_messages", "sms_templates", "scheduled_sms").BindFunc(func(e *core.RecordEnrichEvent) error {
		body := e.Record.GetString("body")
		if decrypted, err := services.DecryptBody(body); err == nil {
			e.Record.Set("body", decrypted)
		}
		return e.Next()
	})
}
