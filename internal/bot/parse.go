package bot

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const dateLayout = "2006-01-02"

var iataRe = regexp.MustCompile(`^[A-Za-z]{3}$`)

type parsedSubscription struct {
	FromIATA  string
	ToIATA    string
	DateFrom  time.Time
	DateTo    time.Time
	Threshold decimal.NullDecimal
}

func parseSubscribeArgs(raw string, now time.Time) (parsedSubscription, error) {
	fields := strings.Fields(raw)
	if len(fields) < 4 || len(fields) > 5 {
		return parsedSubscription{}, errors.New(
			"нужно 4 или 5 аргументов: FROM TO ДАТА_ОТ ДАТА_ДО [ЦЕНА]")
	}

	from, to := strings.ToUpper(fields[0]), strings.ToUpper(fields[1])
	if !iataRe.MatchString(from) || !iataRe.MatchString(to) {
		return parsedSubscription{}, errors.New(
			"IATA-код — это 3 латинские буквы, например MSQ или IST")
	}
	if from == to {
		return parsedSubscription{}, errors.New("город вылета и назначения совпадают")
	}

	dateFrom, err := time.Parse(dateLayout, fields[2])
	if err != nil {
		return parsedSubscription{}, fmt.Errorf("не понял дату %q, нужен формат ГГГГ-ММ-ДД", fields[2])
	}
	dateTo, err := time.Parse(dateLayout, fields[3])
	if err != nil {
		return parsedSubscription{}, fmt.Errorf("не понял дату %q, нужен формат ГГГГ-ММ-ДД", fields[3])
	}
	if dateTo.Before(dateFrom) {
		return parsedSubscription{}, errors.New("дата «до» раньше даты «от»")
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if dateFrom.Before(today) {
		return parsedSubscription{}, errors.New("дата вылета уже в прошлом")
	}

	var threshold decimal.NullDecimal
	if len(fields) == 5 {
		amount, err := decimal.NewFromString(fields[4])
		if err != nil {
			return parsedSubscription{}, fmt.Errorf("не понял цену %q", fields[4])
		}
		if !amount.IsPositive() {
			return parsedSubscription{}, errors.New("цена должна быть больше нуля")
		}
		threshold = decimal.NullDecimal{Decimal: amount, Valid: true}
	}

	return parsedSubscription{
		FromIATA:  from,
		ToIATA:    to,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Threshold: threshold,
	}, nil
}
