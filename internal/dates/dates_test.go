package dates

import (
	"testing"
	"time"
)

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestFormatDay(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "2026-01-01", want: "1 янв"},
		{in: "2026-02-20", want: "20 фев"},
		{in: "2026-12-31", want: "31 дек"},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := FormatDay(date(tc.in)); got != tc.want {
				t.Errorf("FormatDay(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatDayWithYear(t *testing.T) {
	if got, want := FormatDayWithYear(date("2027-01-20")), "20 янв 2027"; got != want {
		t.Errorf("FormatDayWithYear = %q, want %q", got, want)
	}
}

func TestFormatRange(t *testing.T) {
	cases := []struct {
		name string
		from string
		to   string
		want string
	}{
		{name: "same day", from: "2026-02-20", to: "2026-02-20", want: "20 фев"},
		{name: "same month", from: "2026-01-20", to: "2026-01-22", want: "20–22 янв"},
		{name: "same year different months", from: "2026-01-27", to: "2026-02-12", want: "27 янв — 12 фев"},
		{name: "different years", from: "2026-12-28", to: "2027-01-03", want: "28 дек 2026 — 3 янв 2027"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatRange(date(tc.from), date(tc.to)); got != tc.want {
				t.Errorf("FormatRange(%s, %s) = %q, want %q", tc.from, tc.to, got, tc.want)
			}
		})
	}
}
