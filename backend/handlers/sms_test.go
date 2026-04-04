package handlers

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestBuildSendResponse(t *testing.T) {
	t.Run("empty messages", func(t *testing.T) {
		resp := buildSendResponse(nil, 0)

		if resp["batch_id"] != "" {
			t.Errorf("expected empty batch_id, got %q", resp["batch_id"])
		}
		ids := resp["message_ids"].([]string)
		if len(ids) != 0 {
			t.Errorf("expected 0 message_ids, got %d", len(ids))
		}
		if resp["recipients_count"] != 0 {
			t.Errorf("expected recipients_count=0, got %v", resp["recipients_count"])
		}
		if resp["status"] != "accepted" {
			t.Errorf("expected status=accepted, got %v", resp["status"])
		}
	})

	t.Run("with messages", func(t *testing.T) {
		// Create mock records with IDs and batch_id
		collection := &core.Collection{}
		collection.Name = "sms_messages"

		r1 := core.NewRecord(collection)
		r1.Id = "msg_001"
		r1.Set("batch_id", "batch_abc")

		r2 := core.NewRecord(collection)
		r2.Id = "msg_002"
		r2.Set("batch_id", "batch_abc")

		messages := []*core.Record{r1, r2}
		resp := buildSendResponse(messages, 3)

		if resp["batch_id"] != "batch_abc" {
			t.Errorf("expected batch_id=batch_abc, got %v", resp["batch_id"])
		}
		ids := resp["message_ids"].([]string)
		if len(ids) != 2 {
			t.Errorf("expected 2 message_ids, got %d", len(ids))
		}
		if ids[0] != "msg_001" || ids[1] != "msg_002" {
			t.Errorf("unexpected message_ids: %v", ids)
		}
		if resp["recipients_count"] != 3 {
			t.Errorf("expected recipients_count=3, got %v", resp["recipients_count"])
		}
	})
}

func TestE164Regex(t *testing.T) {
	valid := []string{
		"+1234567890",
		"+12345678901234",
		"+5355123456",
		"+442071234567",
	}
	invalid := []string{
		"1234567890",     // missing +
		"+0123456789",    // starts with 0
		"+1",             // too short
		"+123456789012345678", // too long (>15 digits)
		"not-a-number",
		"",
		"+",
		"+ 1234567890",  // space
	}

	for _, v := range valid {
		if !e164Regex.MatchString(v) {
			t.Errorf("expected %q to be valid E.164", v)
		}
	}
	for _, v := range invalid {
		if e164Regex.MatchString(v) {
			t.Errorf("expected %q to be invalid E.164", v)
		}
	}
}
