package services

import (
	"strings"
	"testing"
)

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []string
	}{
		{"no variables", "Hello world", nil},
		{"single variable", "Hello {{name}}", []string{"name"}},
		{"multiple variables", "{{greeting}} {{name}}, code: {{code}}", []string{"greeting", "name", "code"}},
		{"duplicate variables", "{{name}} and {{name}}", []string{"name"}},
		{"reserved and custom", "Hi {{name}}, your code is {{code}}", []string{"name", "code"}},
		{"underscores", "{{first_name}} {{last_name}}", []string{"first_name", "last_name"}},
		{"alphanumeric", "{{var1}} {{var_2}}", []string{"var1", "var_2"}},
		{"invalid: starts with number", "{{1name}}", nil},
		{"invalid: spaces", "{{ name }}", nil},
		{"invalid: special chars", "{{name!}}", nil},
		{"empty braces", "{{}}", nil},
		{"nested braces", "{{{name}}}", []string{"name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractVariables(tt.body)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractVariables(%q) = %v, want %v", tt.body, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExtractVariables(%q)[%d] = %q, want %q", tt.body, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestClassifyVariables(t *testing.T) {
	tests := []struct {
		name         string
		vars         []string
		wantReserved []string
		wantCustom   []string
	}{
		{"empty", nil, nil, nil},
		{"all reserved", []string{"name", "phone"}, []string{"name", "phone"}, nil},
		{"all custom", []string{"code", "url"}, nil, []string{"code", "url"}},
		{"mixed", []string{"name", "code", "phone", "url"}, []string{"name", "phone"}, []string{"code", "url"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reserved, custom := ClassifyVariables(tt.vars)
			if len(reserved) != len(tt.wantReserved) {
				t.Errorf("reserved = %v, want %v", reserved, tt.wantReserved)
			}
			if len(custom) != len(tt.wantCustom) {
				t.Errorf("custom = %v, want %v", custom, tt.wantCustom)
			}
		})
	}
}

func TestInterpolateBody(t *testing.T) {
	tests := []struct {
		name string
		body string
		vars map[string]string
		want string
	}{
		{
			"simple replacement",
			"Hello {{name}}",
			map[string]string{"name": "Alice"},
			"Hello Alice",
		},
		{
			"multiple replacements",
			"{{greeting}} {{name}}, code: {{code}}",
			map[string]string{"greeting": "Hi", "name": "Bob", "code": "1234"},
			"Hi Bob, code: 1234",
		},
		{
			"unresolved variable stays literal",
			"Hello {{name}}, {{unknown}} here",
			map[string]string{"name": "Alice"},
			"Hello Alice, {{unknown}} here",
		},
		{
			"no variables",
			"Hello world",
			map[string]string{"name": "Alice"},
			"Hello world",
		},
		{
			"empty vars map",
			"Hello {{name}}",
			map[string]string{},
			"Hello {{name}}",
		},
		{
			"nil vars map",
			"Hello {{name}}",
			nil,
			"Hello {{name}}",
		},
		{
			"single-pass: no recursive interpolation",
			"Hello {{name}}",
			map[string]string{"name": "{{code}}", "code": "INJECTED"},
			"Hello {{code}}",
		},
		{
			"variable value with braces is literal",
			"Result: {{val}}",
			map[string]string{"val": "{{other}}"},
			"Result: {{other}}",
		},
		{
			"empty value",
			"Hello {{name}}!",
			map[string]string{"name": ""},
			"Hello !",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InterpolateBody(tt.body, tt.vars)
			if got != tt.want {
				t.Errorf("InterpolateBody(%q, %v) = %q, want %q", tt.body, tt.vars, got, tt.want)
			}
		})
	}
}

func TestStripInvisibleUnicode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain text", "Hello world", "Hello world"},
		{"preserves newlines", "Hello\nworld", "Hello\nworld"},
		{"preserves tabs", "Hello\tworld", "Hello\tworld"},
		{"preserves carriage return", "Hello\r\nworld", "Hello\r\nworld"},
		{"strips zero-width space", "Hello\u200Bworld", "Helloworld"},
		{"strips zero-width joiner", "Hello\u200Dworld", "Helloworld"},
		{"strips zero-width non-joiner", "Hello\u200Cworld", "Helloworld"},
		{"strips LTR mark", "Hello\u200Eworld", "Helloworld"},
		{"strips RTL mark", "Hello\u200Fworld", "Helloworld"},
		{"strips soft hyphen", "Hello\u00ADworld", "Helloworld"},
		{"strips BOM", "\uFEFFHello", "Hello"},
		{"strips word joiner", "Hello\u2060world", "Helloworld"},
		{"strips null byte", "Hello\x00world", "Helloworld"},
		{"multiple invisible chars", "\u200B\u200CHello\u200D\u200E", "Hello"},
		{"empty string", "", ""},
		{"only invisible chars", "\u200B\u200C\u200D", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripInvisibleUnicode(tt.in)
			if got != tt.want {
				t.Errorf("StripInvisibleUnicode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestInterpolateBody_LargeTemplate(t *testing.T) {
	body := "Dear {{name}}, " + strings.Repeat("x", 1500) + " {{code}}"
	vars := map[string]string{"name": "Alice", "code": "9999"}
	got := InterpolateBody(body, vars)

	if !strings.HasPrefix(got, "Dear Alice, ") {
		t.Error("expected interpolated prefix")
	}
	if !strings.HasSuffix(got, " 9999") {
		t.Error("expected interpolated suffix")
	}
}
