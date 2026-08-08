package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "0", want: "0"},
		{in: "900.00", want: "900"},
		{in: "1102.37", want: "1 102.37"},
		{in: "1300", want: "1 300"},
		{in: "1234567.5", want: "1 234 567.5"},
		{in: "-1102.37", want: "-1 102.37"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := Format(decimal.RequireFromString(tc.in)); got != tc.want {
				t.Errorf("Format(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatWhole(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "1102.37", want: "1 102"},
		{in: "1102.62", want: "1 103"},
		{in: "999.5", want: "1 000"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := FormatWhole(decimal.RequireFromString(tc.in)); got != tc.want {
				t.Errorf("FormatWhole(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
