package observation

import (
	"context"
	"fmt"
	"strings"

	db "bilecik/pkg"
)

type Repository struct {
	*db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateMany(ctx context.Context, obs []PriceObservation) error {
	if len(obs) == 0 {
		return nil
	}

	rows := make([]string, len(obs))
	args := make([]any, 0, len(obs)*6)

	for i, o := range obs {
		if i == 0 {
			rows[i] = "(?::char(3), ?::char(3), ?::date, ?::numeric, ?::char(3), ?::timestamptz)"
		} else {
			rows[i] = "(?, ?, ?, ?, ?, ?)"
		}
		args = append(args, o.FromIATA, o.ToIATA, o.FlightDate, o.Amount, o.Currency, o.ObservedAt)
	}

	query := fmt.Sprintf(`
		WITH v (from_iata, to_iata, flight_date, amount, currency, observed_at) AS (
			VALUES
				%s
		),
		combined AS (
			SELECT p.from_iata, p.to_iata, p.flight_date, p.amount, p.currency, p.observed_at,
				   false AS is_new
			FROM price_observations p
			JOIN v ON p.from_iata = v.from_iata
				  AND p.to_iata = v.to_iata
				  AND p.flight_date = v.flight_date
			UNION ALL
			SELECT from_iata, to_iata, flight_date, amount, currency, observed_at, true
			FROM v
		),
		marked AS (
			SELECT *,
				   lag(amount)   OVER w AS prev_amount,
				   lag(currency) OVER w AS prev_currency
			FROM combined
			WINDOW w AS (
				PARTITION BY from_iata, to_iata, flight_date
				ORDER BY observed_at, is_new
			)
		)
		INSERT INTO price_observations (id, from_iata, to_iata, flight_date, amount, currency, observed_at)
		SELECT DISTINCT ON (from_iata, to_iata, flight_date, observed_at)
			   gen_random_uuid(), from_iata, to_iata, flight_date, amount, currency, observed_at
		FROM marked
		WHERE is_new
		  AND (prev_amount IS NULL
			   OR prev_amount IS DISTINCT FROM amount
			   OR prev_currency IS DISTINCT FROM currency)
		ORDER BY from_iata, to_iata, flight_date, observed_at`,
		strings.Join(rows, ",\n\t\t\t\t"))

	return repo.DB.WithContext(ctx).Exec(query, args...).Error
}
