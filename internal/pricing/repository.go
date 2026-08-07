package pricing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"bilecik/internal/subscription"
	db "bilecik/pkg"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{db: db}
}

type bestPriceRow struct {
	SubscriptionID uuid.UUID           `gorm:"column:subscription_id"`
	TelegramID     int64               `gorm:"column:telegram_id"`
	FromIATA       string              `gorm:"column:from_iata"`
	ToIATA         string              `gorm:"column:to_iata"`
	DateFrom       time.Time           `gorm:"column:date_from"`
	DateTo         time.Time           `gorm:"column:date_to"`
	Threshold      decimal.NullDecimal `gorm:"column:threshold"`
	CreatedAt      time.Time           `gorm:"column:created_at"`
	PriceRank      sql.NullInt64       `gorm:"column:price_rank"`
	FlightDate     sql.NullTime        `gorm:"column:flight_date"`
	Amount         decimal.NullDecimal `gorm:"column:amount"`
	Currency       sql.NullString      `gorm:"column:currency"`
}

func (r *Repository) ListWithBestPrices(ctx context.Context, topN int) ([]SubscriptionBestPrice, error) {
	if topN < 1 {
		topN = 1
	}

	var rows []bestPriceRow
	err := r.db.WithContext(ctx).Raw(`
		WITH latest_by_date AS (
			SELECT DISTINCT ON (from_iata, to_iata, flight_date)
				from_iata, to_iata, flight_date, amount, currency
			FROM price_observations
			WHERE (from_iata, to_iata) IN (SELECT DISTINCT from_iata, to_iata FROM subscriptions)
			ORDER BY from_iata, to_iata, flight_date, observed_at DESC
		),
		sub_prices AS (
			SELECT s.id AS subscription_id, l.flight_date, ROUND(l.amount) AS amount, l.currency
			FROM subscriptions s
			JOIN latest_by_date l
				ON l.from_iata = s.from_iata
			   AND l.to_iata = s.to_iata
			   AND l.flight_date BETWEEN s.date_from AND s.date_to
		),
		ranked AS (
			SELECT *,
				   DENSE_RANK() OVER (PARTITION BY subscription_id ORDER BY amount ASC) AS price_rank
			FROM sub_prices
		)
		SELECT
			s.id AS subscription_id,
			s.telegram_id,
			s.from_iata,
			s.to_iata,
			s.date_from,
			s.date_to,
			s.threshold,
			s.created_at,
			r.price_rank,
			r.flight_date,
			r.amount,
			r.currency
		FROM subscriptions s
		LEFT JOIN ranked r
			ON r.subscription_id = s.id
		   AND r.price_rank <= ?
		ORDER BY s.created_at DESC, s.id, r.price_rank, r.flight_date`,
		topN).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list with best prices: %w", err)
	}

	return groupRows(rows), nil
}

func groupRows(rows []bestPriceRow) []SubscriptionBestPrice {
	result := make([]SubscriptionBestPrice, 0, len(rows))

	var cur *SubscriptionBestPrice
	var curTier *PriceTier
	var curDates []time.Time

	flushTier := func() {
		if curTier == nil {
			return
		}
		curTier.Dates = compressDates(curDates)
		cur.Tiers = append(cur.Tiers, *curTier)
		curTier, curDates = nil, nil
	}
	flushSub := func() {
		flushTier()
		if cur != nil {
			result = append(result, *cur)
		}
	}

	for _, row := range rows {
		if cur == nil || cur.Subscription.ID != row.SubscriptionID {
			flushSub()
			cur = &SubscriptionBestPrice{Subscription: subscription.Subscription{
				ID:         row.SubscriptionID,
				TelegramID: row.TelegramID,
				FromIATA:   row.FromIATA,
				ToIATA:     row.ToIATA,
				DateFrom:   row.DateFrom,
				DateTo:     row.DateTo,
				Threshold:  row.Threshold,
				CreatedAt:  row.CreatedAt,
			}}
		}

		if !row.PriceRank.Valid {
			continue
		}

		rank := int(row.PriceRank.Int64)
		if curTier == nil || curTier.Rank != rank {
			flushTier()
			curTier = &PriceTier{
				Rank:     rank,
				Amount:   row.Amount.Decimal,
				Currency: row.Currency.String,
			}
		}
		curDates = append(curDates, row.FlightDate.Time)
	}
	flushSub()

	return result
}
