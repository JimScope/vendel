package services

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func makeContact(name, phone string) *core.Record {
	collection := &core.Collection{}
	collection.Name = "contacts"
	r := core.NewRecord(collection)
	r.Set("name", name)
	r.Set("phone_number", phone)
	return r
}

func TestInterpolateForRecipient(t *testing.T) {
	tests := []struct {
		name       string
		tmpl       *TemplateOptions
		phone      string
		contactMap map[string]*core.Record
		want       string
	}{
		{
			"custom variables only",
			&TemplateOptions{
				TemplateBody: "Your code is {{code}}",
				Variables:    map[string]string{"code": "1234"},
			},
			"+5355123456",
			nil,
			"Your code is 1234",
		},
		{
			"reserved variables from contact",
			&TemplateOptions{
				TemplateBody: "Hello {{name}}, your number is {{phone}}",
				Variables:    map[string]string{},
			},
			"+5355123456",
			map[string]*core.Record{
				"+5355123456": makeContact("Alice", "+5355123456"),
			},
			"Hello Alice, your number is +5355123456",
		},
		{
			"mixed custom and reserved",
			&TemplateOptions{
				TemplateBody: "Hi {{name}}, code: {{code}}",
				Variables:    map[string]string{"code": "9999"},
			},
			"+1234567890",
			map[string]*core.Record{
				"+1234567890": makeContact("Bob", "+1234567890"),
			},
			"Hi Bob, code: 9999",
		},
		{
			"contact not in map — reserved vars unresolved",
			&TemplateOptions{
				TemplateBody: "Hi {{name}}, code: {{code}}",
				Variables:    map[string]string{"code": "1111"},
			},
			"+9999999999",
			map[string]*core.Record{
				"+1234567890": makeContact("Bob", "+1234567890"),
			},
			"Hi {{name}}, code: 1111",
		},
		{
			"nil contact map",
			&TemplateOptions{
				TemplateBody: "Hello {{name}}",
				Variables:    map[string]string{},
			},
			"+5355123456",
			nil,
			"Hello {{name}}",
		},
		{
			"reserved vars override user-supplied keys",
			&TemplateOptions{
				TemplateBody: "Hello {{name}}",
				Variables:    map[string]string{"name": "INJECTED"},
			},
			"+5355123456",
			map[string]*core.Record{
				"+5355123456": makeContact("Alice", "+5355123456"),
			},
			"Hello Alice",
		},
		{
			"strips invisible unicode from result",
			&TemplateOptions{
				TemplateBody: "Hello {{name}}",
				Variables:    map[string]string{"name": "Alice\u200B"},
			},
			"+5355123456",
			nil,
			"Hello Alice",
		},
		{
			"empty template body",
			&TemplateOptions{
				TemplateBody: "",
				Variables:    map[string]string{"code": "1234"},
			},
			"+5355123456",
			nil,
			"",
		},
		{
			"no variables in template",
			&TemplateOptions{
				TemplateBody: "Static message",
				Variables:    map[string]string{},
			},
			"+5355123456",
			nil,
			"Static message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpolateForRecipient(tt.tmpl, tt.phone, tt.contactMap)
			if got != tt.want {
				t.Errorf("interpolateForRecipient() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInterpolateForRecipient_PerRecipientDifference(t *testing.T) {
	tmpl := &TemplateOptions{
		TemplateBody: "Hi {{name}}, your code is {{code}}",
		Variables:    map[string]string{"code": "SALE20"},
	}
	contactMap := map[string]*core.Record{
		"+1111111111": makeContact("Alice", "+1111111111"),
		"+2222222222": makeContact("Bob", "+2222222222"),
	}

	got1 := interpolateForRecipient(tmpl, "+1111111111", contactMap)
	got2 := interpolateForRecipient(tmpl, "+2222222222", contactMap)

	if got1 == got2 {
		t.Error("expected different messages per recipient")
	}
	if got1 != "Hi Alice, your code is SALE20" {
		t.Errorf("recipient 1: got %q", got1)
	}
	if got2 != "Hi Bob, your code is SALE20" {
		t.Errorf("recipient 2: got %q", got2)
	}
}
