package services

import "testing"

func TestFoldToGSM7(t *testing.T) {
	cases := map[string]string{
		// Decomposed accent (U+0301), as typed in the 2026-11-26 template.
		"invitación al ATS que ganará": "invitacion al ATS que ganara",
		"Hola José, María y Ángel":       "Hola José, Maria y Angel",
		"¿Señor Núñez? ¡Gracias!":        "¿Señor Nuñez? ¡Gracias!",
		"“Hola” — prueba…":               "\"Hola\" - prueba...",
		"Entrevista 🎉":                   "Entrevista ?",
		"plain ascii 123":                "plain ascii 123",
	}
	for in, want := range cases {
		if got := FoldToGSM7(in); got != want {
			t.Errorf("FoldToGSM7(%q) = %q, want %q", in, got, want)
		}
		for _, r := range FoldToGSM7(in) {
			if !isGSM7(r) {
				t.Errorf("FoldToGSM7(%q) kept non-GSM rune %q", in, r)
			}
		}
	}
}
