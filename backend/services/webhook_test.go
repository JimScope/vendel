package services

import (
	"testing"
	"time"
)

func TestJitteredBackoff_WithinBounds(t *testing.T) {
	// Every jittered value must land inside [base*(1-jitter), base*(1+jitter))
	// and stay strictly positive, so the retry schedule keeps growing and never
	// fires immediately. Sampled many times because the perturbation is random.
	for _, base := range WebhookRetryBackoffs {
		lo := time.Duration(float64(base) * (1 - WebhookRetryJitter))
		hi := time.Duration(float64(base) * (1 + WebhookRetryJitter))
		for i := 0; i < 10000; i++ {
			got := jitteredBackoff(base)
			if got <= 0 {
				t.Fatalf("jitteredBackoff(%v) = %v, must stay positive", base, got)
			}
			if got < lo || got >= hi {
				t.Fatalf("jitteredBackoff(%v) = %v, want within [%v, %v)", base, got, lo, hi)
			}
		}
	}
}

func TestWebhookHost_LowercasesHostname(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://Example.com/path", "example.com"},
		{"https://EXAMPLE.COM/path", "example.com"},
		{"https://example.com/path", "example.com"},
		{"https://Api.Example.com:8443/x", "api.example.com"},
	}
	for _, tc := range cases {
		got := webhookHost(tc.in)
		if got != tc.want {
			t.Errorf("webhookHost(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWebhookHost_UnparseableReturnsSentinel(t *testing.T) {
	cases := []string{
		"",
		"://broken",
		"not a url",
		"http:///no-host",
	}
	for _, in := range cases {
		got := webhookHost(in)
		if got != webhookUnparseableHostBucket {
			t.Errorf("webhookHost(%q) = %q, want sentinel %q", in, got, webhookUnparseableHostBucket)
		}
	}
}
