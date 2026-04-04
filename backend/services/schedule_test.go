package services

import (
	"testing"
	"time"
)

func TestValidateCronExpression(t *testing.T) {
	valid := []string{
		"* * * * *",       // every minute
		"0 9 * * *",       // 9am daily
		"0 0 1 * *",       // 1st of each month
		"*/5 * * * *",     // every 5 minutes
		"0 9 * * 1-5",     // weekdays at 9am
		"30 14 * * *",     // 2:30pm daily
	}
	invalid := []string{
		"",
		"not-a-cron",
		"* * *",            // too few fields
		"60 * * * *",       // invalid minute
		"* 25 * * *",       // invalid hour
	}

	for _, expr := range valid {
		if err := ValidateCronExpression(expr); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", expr, err)
		}
	}
	for _, expr := range invalid {
		if err := ValidateCronExpression(expr); err == nil {
			t.Errorf("expected %q to be invalid", expr)
		}
	}
}

func TestComputeNextRun(t *testing.T) {
	t.Run("valid cron returns future time", func(t *testing.T) {
		result, err := ComputeNextRun("* * * * *", "UTC")
		if err != nil {
			t.Fatal(err)
		}

		parsed, err := time.Parse(time.RFC3339, result)
		if err != nil {
			t.Fatalf("invalid RFC3339: %v", err)
		}

		if !parsed.After(time.Now().UTC().Add(-time.Minute)) {
			t.Error("next run should be in the future")
		}
	})

	t.Run("respects timezone", func(t *testing.T) {
		// 9am daily — result should differ by timezone offset
		resultUTC, err := ComputeNextRun("0 9 * * *", "UTC")
		if err != nil {
			t.Fatal(err)
		}
		resultNY, err := ComputeNextRun("0 9 * * *", "America/New_York")
		if err != nil {
			t.Fatal(err)
		}

		// The UTC results should differ because 9am UTC != 9am ET
		if resultUTC == resultNY {
			t.Error("expected different UTC results for different timezones")
		}
	})

	t.Run("invalid cron returns error", func(t *testing.T) {
		_, err := ComputeNextRun("invalid", "UTC")
		if err == nil {
			t.Error("expected error for invalid cron")
		}
	})

	t.Run("invalid timezone returns error", func(t *testing.T) {
		_, err := ComputeNextRun("* * * * *", "Not/A/Timezone")
		if err == nil {
			t.Error("expected error for invalid timezone")
		}
	})

	t.Run("result is RFC3339 UTC", func(t *testing.T) {
		result, err := ComputeNextRun("0 12 * * *", "America/Chicago")
		if err != nil {
			t.Fatal(err)
		}

		parsed, err := time.Parse(time.RFC3339, result)
		if err != nil {
			t.Fatalf("should be valid RFC3339: %v", err)
		}

		if parsed.Location() != time.UTC {
			t.Error("result should be in UTC")
		}
	})
}
