package pricing

import (
	"sort"
	"time"

	"bilecik/internal/subscription"

	"github.com/shopspring/decimal"
)

type DateRange struct {
	From time.Time
	To   time.Time
}

type PriceTier struct {
	Rank     int
	Amount   decimal.Decimal
	Currency string
	Dates    []DateRange
}

type SubscriptionBestPrice struct {
	Subscription subscription.Subscription
	Tiers        []PriceTier
}

func compressDates(dates []time.Time) []DateRange {
	if len(dates) == 0 {
		return nil
	}

	normalized := make([]time.Time, len(dates))
	for i, d := range dates {
		normalized[i] = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Before(normalized[j]) })

	ranges := make([]DateRange, 0, len(normalized))
	start, prev := normalized[0], normalized[0]
	for _, d := range normalized[1:] {
		if d.Sub(prev) > 24*time.Hour {
			ranges = append(ranges, DateRange{From: start, To: prev})
			start = d
		}
		prev = d
	}
	ranges = append(ranges, DateRange{From: start, To: prev})

	return ranges
}
