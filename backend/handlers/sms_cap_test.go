package handlers

import (
	"fmt"
	"strings"
	"testing"

	"vendel/services"

	"github.com/pocketbase/pocketbase/core"
)

// TestResolveRecipientsCapDirect verifies the recipient cap (REL-3) rejects a
// request whose direct recipient list exceeds MaxRecipientsPerRequest, and that
// a request exactly at the cap is accepted.
func TestResolveRecipientsCapDirect(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	t.Run("over cap rejected", func(t *testing.T) {
		recipients := make([]string, services.MaxRecipientsPerRequest+1)
		for i := range recipients {
			recipients[i] = fmt.Sprintf("+1%09d", 100000000+i) // unique, valid E.164
		}
		_, err := resolveRecipients(app, "someuser", recipients, nil)
		if err == nil || !strings.Contains(err.Error(), "too many recipients") {
			t.Fatalf("expected too-many-recipients error, got %v", err)
		}
	})

	t.Run("exactly at cap accepted", func(t *testing.T) {
		recipients := make([]string, services.MaxRecipientsPerRequest)
		for i := range recipients {
			recipients[i] = fmt.Sprintf("+1%09d", 100000000+i)
		}
		got, err := resolveRecipients(app, "someuser", recipients, nil)
		if err != nil {
			t.Fatalf("expected no error at cap, got %v", err)
		}
		if len(got) != services.MaxRecipientsPerRequest {
			t.Fatalf("expected %d recipients, got %d", services.MaxRecipientsPerRequest, len(got))
		}
	})
}

// TestResolveRecipientsCapGroupExpansion verifies the cap also covers group
// expansion (REL-3): a group that expands beyond MaxRecipientsPerRequest must
// be rejected before any quota is reserved.
func TestResolveRecipientsCapGroupExpansion(t *testing.T) {
	app := setupTestApp(t)
	defer app.Cleanup()

	user, err := app.FindAuthRecordByEmail("users", "user@test.com")
	if err != nil {
		t.Fatalf("failed to find test user: %v", err)
	}

	groupsCol, err := app.FindCollectionByNameOrId("contact_groups")
	if err != nil {
		t.Fatal(err)
	}
	group := core.NewRecord(groupsCol)
	group.Set("user", user.Id)
	group.Set("name", "BigGroup")
	if err := app.Save(group); err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	contactsCol, err := app.FindCollectionByNameOrId("contacts")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < services.MaxRecipientsPerRequest+1; i++ {
		c := core.NewRecord(contactsCol)
		c.Set("user", user.Id)
		c.Set("name", fmt.Sprintf("Contact %d", i))
		c.Set("phone_number", fmt.Sprintf("+1%09d", 200000000+i))
		c.Set("groups", []string{group.Id})
		if err := app.Save(c); err != nil {
			t.Fatalf("failed to create contact %d: %v", i, err)
		}
	}

	_, err = resolveRecipients(app, user.Id, nil, []string{group.Id})
	if err == nil || !strings.Contains(err.Error(), "too many recipients") {
		t.Fatalf("expected too-many-recipients error from group expansion, got %v", err)
	}
}
