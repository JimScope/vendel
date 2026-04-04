package handlers

import (
	"strings"
	"testing"
)

func TestParseVCard(t *testing.T) {
	t.Run("single contact", func(t *testing.T) {
		vcf := `BEGIN:VCARD
VERSION:3.0
FN:Alice Smith
TEL:+1234567890
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].name != "Alice Smith" {
			t.Errorf("name = %q, want %q", entries[0].name, "Alice Smith")
		}
		if entries[0].phone != "+1234567890" {
			t.Errorf("phone = %q, want %q", entries[0].phone, "+1234567890")
		}
	})

	t.Run("multiple contacts", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:Alice
TEL:+1111111111
END:VCARD
BEGIN:VCARD
FN:Bob
TEL:+2222222222
END:VCARD
BEGIN:VCARD
FN:Charlie
TEL:+3333333333
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}
		if entries[2].name != "Charlie" {
			t.Errorf("third entry name = %q", entries[2].name)
		}
	})

	t.Run("TEL with TYPE parameter", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:Alice
TEL;TYPE=CELL:+1234567890
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].phone != "+1234567890" {
			t.Errorf("phone = %q, want +1234567890", entries[0].phone)
		}
	})

	t.Run("TEL with uri format", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:Alice
TEL;VALUE=uri:tel:+1234567890
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].phone != "+1234567890" {
			t.Errorf("phone = %q, want +1234567890", entries[0].phone)
		}
	})

	t.Run("phone with formatting chars stripped", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:Alice
TEL:+1 (234) 567-8901
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].phone != "+12345678901" {
			t.Errorf("phone = %q, want +12345678901", entries[0].phone)
		}
	})

	t.Run("skip contact without phone", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:No Phone
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("skip contact without name", func(t *testing.T) {
		vcf := `BEGIN:VCARD
TEL:+1234567890
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("skip invalid phone number", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:Bad Phone
TEL:not-a-number
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		entries := parseVCard(strings.NewReader(""))
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("first phone wins", func(t *testing.T) {
		vcf := `BEGIN:VCARD
FN:Alice
TEL:+1111111111
TEL:+2222222222
END:VCARD`

		entries := parseVCard(strings.NewReader(vcf))
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].phone != "+1111111111" {
			t.Errorf("should use first phone, got %q", entries[0].phone)
		}
	})
}
