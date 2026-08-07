package pricing

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}

func nullInt64(v int64) sql.NullInt64    { return sql.NullInt64{Int64: v, Valid: true} }
func nullTime(v time.Time) sql.NullTime  { return sql.NullTime{Time: v, Valid: true} }
func nullString(v string) sql.NullString { return sql.NullString{String: v, Valid: true} }
func nullDecimal(v string) decimal.NullDecimal {
	return decimal.NullDecimal{Decimal: decimal.RequireFromString(v), Valid: true}
}

func TestCompressDates(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []DateRange
	}{
		{
			name: "empty",
			in:   nil,
			want: nil,
		},
		{
			name: "single day",
			in:   []string{"2026-01-27"},
			want: []DateRange{{From: date("2026-01-27"), To: date("2026-01-27")}},
		},
		{
			name: "contiguous run",
			in:   []string{"2026-01-25", "2026-01-26", "2026-01-27"},
			want: []DateRange{{From: date("2026-01-25"), To: date("2026-01-27")}},
		},
		{
			name: "gap splits into two ranges",
			in:   []string{"2026-01-01", "2026-01-02", "2026-01-10"},
			want: []DateRange{
				{From: date("2026-01-01"), To: date("2026-01-02")},
				{From: date("2026-01-10"), To: date("2026-01-10")},
			},
		},
		{
			name: "unsorted input still sorted",
			in:   []string{"2026-01-10", "2026-01-01", "2026-01-02"},
			want: []DateRange{
				{From: date("2026-01-01"), To: date("2026-01-02")},
				{From: date("2026-01-10"), To: date("2026-01-10")},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dates := make([]time.Time, len(tc.in))
			for i, s := range tc.in {
				dates[i] = date(s)
			}

			got := compressDates(dates)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d ranges, want %d: %+v", len(got), len(tc.want), got)
			}
			for i := range got {
				if !got[i].From.Equal(tc.want[i].From) || !got[i].To.Equal(tc.want[i].To) {
					t.Errorf("range %d = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestGroupRows(t *testing.T) {
	sub1 := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	sub2 := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	rows := []bestPriceRow{
		{SubscriptionID: sub1, FromIATA: "MSQ", ToIATA: "IST", PriceRank: nullInt64(1), FlightDate: nullTime(date("2026-01-25")), Amount: nullDecimal("100"), Currency: nullString("BYN")},
		{SubscriptionID: sub1, FromIATA: "MSQ", ToIATA: "IST", PriceRank: nullInt64(1), FlightDate: nullTime(date("2026-01-26")), Amount: nullDecimal("100"), Currency: nullString("BYN")},
		{SubscriptionID: sub1, FromIATA: "MSQ", ToIATA: "IST", PriceRank: nullInt64(2), FlightDate: nullTime(date("2026-01-30")), Amount: nullDecimal("120"), Currency: nullString("BYN")},
		{SubscriptionID: sub2, FromIATA: "MSQ", ToIATA: "LED"},
	}

	got := groupRows(rows)
	if len(got) != 2 {
		t.Fatalf("got %d subscriptions, want 2", len(got))
	}

	if got[0].Subscription.ID != sub1 {
		t.Fatalf("first subscription id = %v, want %v", got[0].Subscription.ID, sub1)
	}
	if len(got[0].Tiers) != 2 {
		t.Fatalf("got %d tiers, want 2: %+v", len(got[0].Tiers), got[0].Tiers)
	}
	if len(got[0].Tiers[0].Dates) != 1 || got[0].Tiers[0].Dates[0].From != date("2026-01-25") || got[0].Tiers[0].Dates[0].To != date("2026-01-26") {
		t.Errorf("tier 1 dates = %+v", got[0].Tiers[0].Dates)
	}
	if got[0].Tiers[1].Rank != 2 || got[0].Tiers[1].Amount.String() != "120" {
		t.Errorf("tier 2 = %+v", got[0].Tiers[1])
	}

	if got[1].Subscription.ID != sub2 {
		t.Fatalf("second subscription id = %v, want %v", got[1].Subscription.ID, sub2)
	}
	if len(got[1].Tiers) != 0 {
		t.Errorf("subscription with no observations should have no tiers, got %+v", got[1].Tiers)
	}
}
