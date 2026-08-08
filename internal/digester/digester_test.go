package digester

import (
	"testing"
	"time"

	"bilecik/internal/pricing"
	"bilecik/internal/subscription"

	"github.com/shopspring/decimal"
)

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestSubscriptionHeader(t *testing.T) {
	cases := []struct {
		name string
		sub  subscription.Subscription
		want string
	}{
		{
			name: "same year appends year once",
			sub:  subscription.Subscription{FromIATA: "MSQ", ToIATA: "IST", DateFrom: date("2026-01-20"), DateTo: date("2026-01-25")},
			want: "✈️ MSQ → IST · 20–25 янв 2026",
		},
		{
			name: "cross-year keeps both years from range",
			sub:  subscription.Subscription{FromIATA: "MSQ", ToIATA: "LED", DateFrom: date("2026-12-28"), DateTo: date("2027-01-03")},
			want: "✈️ MSQ → LED · 28 дек 2026 – 3 янв 2027",
		},
		{
			name: "cities render as city (iata)",
			sub:  subscription.Subscription{FromIATA: "MSQ", FromCity: "Минск", ToIATA: "IST", ToCity: "Стамбул", DateFrom: date("2026-01-20"), DateTo: date("2026-01-25")},
			want: "✈️ Минск (MSQ) → Стамбул (IST) · 20–25 янв 2026",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := subscriptionHeader(tc.sub); got != tc.want {
				t.Errorf("subscriptionHeader = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatTier(t *testing.T) {
	cases := []struct {
		name string
		tier pricing.PriceTier
		want string
	}{
		{
			name: "rank 1 gets medal",
			tier: pricing.PriceTier{
				Rank:     1,
				Amount:   decimal.RequireFromString("100"),
				Currency: "BYN",
				Dates:    []pricing.DateRange{{From: date("2026-01-25"), To: date("2026-01-26")}},
			},
			want: "🥇 100 BYN — 25–26 янв",
		},
		{
			name: "rank 3 gets medal",
			tier: pricing.PriceTier{
				Rank:     3,
				Amount:   decimal.RequireFromString("140"),
				Currency: "BYN",
				Dates:    []pricing.DateRange{{From: date("2026-02-01"), To: date("2026-02-01")}},
			},
			want: "🥉 140 BYN — 1 фев",
		},
		{
			name: "rank beyond medals falls back to number",
			tier: pricing.PriceTier{
				Rank:     4,
				Amount:   decimal.RequireFromString("150"),
				Currency: "BYN",
				Dates:    []pricing.DateRange{{From: date("2026-02-02"), To: date("2026-02-02")}},
			},
			want: "4) 150 BYN — 2 фев",
		},
		{
			name: "multiple ranges joined",
			tier: pricing.PriceTier{
				Rank:     1,
				Amount:   decimal.RequireFromString("100"),
				Currency: "BYN",
				Dates: []pricing.DateRange{
					{From: date("2026-01-25"), To: date("2026-01-26")},
					{From: date("2026-01-30"), To: date("2026-01-30")},
				},
			},
			want: "🥇 100 BYN — 25–26 янв · 30 янв",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatTier(tc.tier); got != tc.want {
				t.Errorf("formatTier = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatDailyDigest(t *testing.T) {
	sub := subscription.Subscription{FromIATA: "MSQ", ToIATA: "IST", DateFrom: date("2026-01-20"), DateTo: date("2026-02-10")}

	t.Run("with tiers", func(t *testing.T) {
		prices := []pricing.SubscriptionBestPrice{
			{
				Subscription: sub,
				Tiers: []pricing.PriceTier{
					{
						Rank:     1,
						Amount:   decimal.RequireFromString("100"),
						Currency: "BYN",
						Dates:    []pricing.DateRange{{From: date("2026-01-25"), To: date("2026-01-26")}},
					},
					{
						Rank:     2,
						Amount:   decimal.RequireFromString("120"),
						Currency: "BYN",
						Dates:    []pricing.DateRange{{From: date("2026-01-30"), To: date("2026-01-30")}},
					},
				},
			},
		}

		want := "📊 Свежая сводка по твоим подпискам\n" +
			"\n" +
			"✈️ MSQ → IST · 20 янв – 10 фев 2026\n" +
			"🥇 100 BYN — 25–26 янв\n" +
			"🥈 120 BYN — 30 янв"
		if got := formatDailyDigest(prices); got != want {
			t.Errorf("formatDailyDigest =\n%q\nwant\n%q", got, want)
		}
	})

	t.Run("subscription without prices", func(t *testing.T) {
		prices := []pricing.SubscriptionBestPrice{{Subscription: sub}}

		want := "📊 Свежая сводка по твоим подпискам\n" +
			"\n" +
			"✈️ MSQ → IST · 20 янв – 10 фев 2026\n" +
			"⏳ пока нет данных о ценах"
		if got := formatDailyDigest(prices); got != want {
			t.Errorf("formatDailyDigest =\n%q\nwant\n%q", got, want)
		}
	})
}
