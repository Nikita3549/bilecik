package bot

import (
	"testing"
)

func TestValidateIATA(t *testing.T) {
	t.Run("valid lowercased", func(t *testing.T) {
		got, err := validateIATA("  msq ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "MSQ" {
			t.Errorf("want MSQ, got %q", got)
		}
	})

	for _, in := range []string{"", "MS", "MSQQ", "M1Q", "МСК"} {
		t.Run("invalid "+in, func(t *testing.T) {
			if _, err := validateIATA(in); err == nil {
				t.Errorf("expected error for %q", in)
			}
		})
	}
}

func TestParseFlightDate(t *testing.T) {
	if _, err := parseFlightDate("2026-08-01"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, in := range []string{"2026/08/01", "01-08-2026", "", "tomorrow"} {
		if _, err := parseFlightDate(in); err == nil {
			t.Errorf("expected error for %q", in)
		}
	}
}

func TestParseThreshold(t *testing.T) {
	t.Run("dash skips", func(t *testing.T) {
		got, err := parseThreshold("-")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Valid {
			t.Errorf("dash must yield empty threshold")
		}
	})

	t.Run("positive value", func(t *testing.T) {
		got, err := parseThreshold("250.5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Valid || got.Decimal.String() != "250.5" {
			t.Errorf("bad parse: %+v", got)
		}
	})

	for _, in := range []string{"0", "-5", "cheap", ""} {
		if _, err := parseThreshold(in); err == nil {
			t.Errorf("expected error for %q", in)
		}
	}
}
