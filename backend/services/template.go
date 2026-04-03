package services

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// variableRegex matches {{variable_name}} placeholders.
	variableRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)

	// ReservedVariables are auto-resolved from contact data per recipient.
	ReservedVariables = map[string]bool{
		"name":  true,
		"phone": true,
	}
)

// ExtractVariables returns unique variable names found in a template body.
func ExtractVariables(body string) []string {
	matches := variableRegex.FindAllStringSubmatch(body, -1)
	seen := make(map[string]bool)
	var vars []string
	for _, m := range matches {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}
	return vars
}

// InterpolateBody replaces {{var}} placeholders with values from the map.
// Single-pass only — the result is never re-processed, preventing injection.
// Unresolved variables are left as-is (e.g. {{unknown}} stays literal).
func InterpolateBody(body string, vars map[string]string) string {
	return variableRegex.ReplaceAllStringFunc(body, func(match string) string {
		name := match[2 : len(match)-2] // strip {{ and }}
		if val, ok := vars[name]; ok {
			return val
		}
		return match
	})
}

// StripInvisibleUnicode removes invisible and formatting unicode characters
// that can manipulate perceived message length. Preserves newlines, tabs, and spaces.
func StripInvisibleUnicode(s string) string {
	return strings.Map(func(r rune) rune {
		// Keep printable ASCII and common whitespace
		if r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		// Remove Unicode format characters (Cf category)
		// Includes: zero-width space (U+200B), ZWJ/ZWNJ (U+200C/D),
		// LTR/RTL marks (U+200E/F, U+202A-202E, U+2066-2069),
		// BOM (U+FEFF), soft hyphen (U+00AD), word joiner (U+2060)
		if unicode.Is(unicode.Cf, r) {
			return -1
		}
		// Remove control characters (Cc) except the whitespace we kept above
		if unicode.Is(unicode.Cc, r) {
			return -1
		}
		return r
	}, s)
}

// ClassifyVariables splits variable names into reserved (auto-filled from contact)
// and custom (user-provided at send time).
func ClassifyVariables(vars []string) (reserved, custom []string) {
	for _, v := range vars {
		if ReservedVariables[v] {
			reserved = append(reserved, v)
		} else {
			custom = append(custom, v)
		}
	}
	return
}
