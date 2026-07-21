package bot

import (
	"testing"
	"time"
)

func TestParseSubscribeArgs(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)

	t.Run("valid without threshold", func(t *testing.T) {
		got, err := parseSubscribeArgs("msq ist 2026-08-01 2026-08-10", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FromIATA != "MSQ" || got.ToIATA != "IST" {
			t.Errorf("iata not uppercased: %+v", got)
		}
		if got.Threshold.Valid {
			t.Errorf("threshold should be absent")
		}
	})

	t.Run("valid with threshold", func(t *testing.T) {
		got, err := parseSubscribeArgs("MSQ IST 2026-08-01 2026-08-10 250.5", now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Threshold.Valid || got.Threshold.Decimal.String() != "250.5" {
			t.Errorf("threshold parsed wrong: %+v", got.Threshold)
		}
	})

	cases := []struct {
		name string
		args string
	}{
		{"too few args", "MSQ IST 2026-08-01"},
		{"too many args", "MSQ IST 2026-08-01 2026-08-10 250 extra"},
		{"bad iata length", "MSQQ IST 2026-08-01 2026-08-10"},
		{"digits in iata", "M1Q IST 2026-08-01 2026-08-10"},
		{"same city", "MSQ MSQ 2026-08-01 2026-08-10"},
		{"bad date format", "MSQ IST 2026/08/01 2026-08-10"},
		{"reversed dates", "MSQ IST 2026-08-10 2026-08-01"},
		{"date in past", "MSQ IST 2026-01-01 2026-08-10"},
		{"negative threshold", "MSQ IST 2026-08-01 2026-08-10 -5"},
		{"non-numeric threshold", "MSQ IST 2026-08-01 2026-08-10 cheap"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseSubscribeArgs(tc.args, now); err == nil {
				t.Errorf("expected error for %q, got nil", tc.args)
			}
		})
	}
}
