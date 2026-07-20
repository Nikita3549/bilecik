package belavia

import (
	"time"

	"github.com/shopspring/decimal"
)

type Observation struct {
	From       string
	To         string
	FlightDate time.Time
	Amount     decimal.Decimal
	Currency   string
	ObservedAt time.Time
}

type apiResponse struct {
	Data struct {
		FlightsMinPricesInPeriodRT struct {
			DatesWithLowestPricesByLegs []legDates `json:"datesWithLowestPricesByLegs"`
		} `json:"FlightsMinPricesInPeriodRT"`
	} `json:"data"`
}

type legDates struct {
	LegID  string          `json:"legId"`
	OneWay []dateWithPrice `json:"oneWay"`
}

type dateWithPrice struct {
	Date  string `json:"date"`
	Price *money `json:"price"`
}

type money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}

type graphQLRequest struct {
	OperationName string      `json:"operationName"`
	Variables     requestVars `json:"variables"`
	Query         string      `json:"query"`
}

type requestVars struct {
	Parameters requestParams `json:"parameters"`
}

type requestParams struct {
	PromotionCode *string     `json:"promotionCode"`
	Currency      string      `json:"currency"`
	Passengers    []passenger `json:"passengers"`
	Segments      []segment   `json:"segments"`
	FFPMode       bool        `json:"ffpMode"`
	DaysCount     int         `json:"daysCount"`
}

type passenger struct {
	Count                 int     `json:"count"`
	PassengerType         string  `json:"passengerType"`
	ExtendedPassengerType *string `json:"extendedPassengerType"`
}

type segment struct {
	Date      string  `json:"date"`
	Departure iataRef `json:"departure"`
	Arrival   iataRef `json:"arrival"`
}

type iataRef struct {
	IATA string `json:"iata"`
}
