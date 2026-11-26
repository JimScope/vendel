package services

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// gsm7Basic and gsm7Extension are the characters of the GSM 03.38 default
// alphabet. A body made only of them fits 160 characters per SMS; any other
// character forces UCS-2 and a 70-character limit.
const gsm7Basic = "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?" +
	"¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà"
const gsm7Extension = "^{}\\[~]|€\f"

var gsm7Replacements = map[rune]string{
	'‘': "'", '’': "'", '‚': "'", '´': "'", '`': "'",
	'“': "\"", '”': "\"", '„': "\"", '«': "\"", '»': "\"",
	'–': "-", '—': "-", '‐': "-", '−': "-",
	'…': "...", ' ': " ", '\t': " ", '•': "*", '·': ".",
}

func isGSM7(r rune) bool {
	return strings.ContainsRune(gsm7Basic, r) || strings.ContainsRune(gsm7Extension, r)
}

// FoldToGSM7 rewrites a body into the GSM 7-bit alphabet so an Android agent
// can send it as one plain SMS (up to 160 characters). Characters the alphabet
// already has are kept (ñ, é, ü, ¿, ¡); accented letters it lacks lose their
// accent (á → a, ó → o); typographic punctuation becomes its ASCII form; and
// anything else becomes '?'.
//
// The Android agent sends with a single-part API, so a UCS-2 body over 70
// characters fails on the phone with "Generic failure" (2026-11-26 incident,
// docs/android-agent-delivery-plan.md).
func FoldToGSM7(s string) string {
	var b strings.Builder
	for _, r := range norm.NFC.String(s) {
		switch {
		case isGSM7(r):
			b.WriteRune(r)
		case gsm7Replacements[r] != "":
			b.WriteString(gsm7Replacements[r])
		default:
			folded := false
			for _, base := range norm.NFD.String(string(r)) {
				if unicode.Is(unicode.Mn, base) {
					continue
				}
				if isGSM7(base) {
					b.WriteRune(base)
					folded = true
				}
			}
			if !folded {
				b.WriteRune('?')
			}
		}
	}
	return b.String()
}
