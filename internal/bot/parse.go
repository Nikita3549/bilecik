package bot

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const dateLayout = "2006-01-02"

var iataRe = regexp.MustCompile(`^[A-Za-z]{3}$`)

func validateIATA(s string) (string, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if !iataRe.MatchString(s) {
		return "", errors.New("IATA-код — это 3 латинские буквы, например MSQ или IST")
	}
	return s, nil
}

func parseFlightDate(s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, errors.New("нужен формат даты ГГГГ-ММ-ДД, например 2026-08-01")
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
