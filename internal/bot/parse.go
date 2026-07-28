package bot

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const dateLayout = "02.01.2006"

func parseFlightDate(s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, errors.New("нужен формат даты ДД.ММ.ГГГГ, например 01.08.2026")
	}
	return t, nil
}

func parseThreshold(s string) (decimal.NullDecimal, error) {
	s = strings.TrimSpace(s)
	if s == "-" {
		return decimal.NullDecimal{}, nil
	}
	amount, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.NullDecimal{}, errors.New("цена — это число, например 250, или «-» чтобы пропустить")
	}
	if !amount.IsPositive() {
		return decimal.NullDecimal{}, errors.New("цена должна быть больше нуля")
	}
	return decimal.NullDecimal{Decimal: amount, Valid: true}, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
