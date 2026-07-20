package belavia

import (
	"encoding/json"
	"fmt"
	"time"
)

const flightsMinPricesQuery = `query FlightsMinPricesInPeriodRT($parameters: FlightsMinPricesInPeriodParameters!) {
  FlightsMinPricesInPeriodRT(parameters: $parameters) {
    datesWithLowestPricesByLegs {
      ...DateWithPriceByLegs
      __typename
    }
    __typename
  }
}

fragment DateWithPriceByLegs on DateWithPriceByLegs {
  legId
  oneWay {
    ...DateWithPrice
    __typename
  }
  halfRoundTrip {
    ...DateWithPrice
    __typename
  }
  __typename
}

fragment DateWithPrice on DateWithPrice {
  date
  price {
    ...Money
    __typename
  }
  info
  __typename
}

fragment Money on Money {
  amount
  currency
  __typename
}
`
const dateLayout = "2006-01-02"

func decodeBelaviaResponse(body []byte) (apiResponse, error) {
	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apiResponse{}, fmt.Errorf("decode belavia response %v", err)
	}
	return resp, nil
}

func buildOneWayRequest(from, to string, startDate time.Time) graphQLRequest {
	return graphQLRequest{
		OperationName: "FlightsMinPricesInPeriodRT",
		Query:         flightsMinPricesQuery,
		Variables: requestVars{
			Parameters: requestParams{
				PromotionCode: nil,
				Currency:      "BYN",
				Passengers: []passenger{
					{Count: 1, PassengerType: "ADT"},
					{Count: 0, PassengerType: "CLD"},
					{Count: 0, PassengerType: "INF"},
				},
				Segments: []segment{
					{
						Date:      startDate.Format(dateLayout),
						Departure: iataRef{IATA: from},
						Arrival:   iataRef{IATA: to},
					},
					{
						Date:      startDate.Format(dateLayout),
						Departure: iataRef{IATA: to},
						Arrival:   iataRef{IATA: from},
					},
				},
				FFPMode:   false,
				DaysCount: apiDaysCount,
			},
		},
	}
}

func parseObservations(raw apiResponse, from, to string, observedAt time.Time) ([]Observation, error) {
	legs := raw.Data.FlightsMinPricesInPeriodRT.DatesWithLowestPricesByLegs

	if len(legs) == 0 {
		return nil, nil
	}

	dates := legs[0].OneWay
	observations := make([]Observation, 0, len(dates))

	for _, d := range dates {
		if d.Price == nil {
			continue
		}

		flightDate, err := time.Parse(dateLayout, d.Date)
		if err != nil {
			return nil, fmt.Errorf("parse flight date %q: %w", d.Date, err)
		}

		observations = append(observations, Observation{
			From:       from,
			To:         to,
			FlightDate: flightDate,
			Amount:     d.Price.Amount,
			Currency:   d.Price.Currency,
			ObservedAt: observedAt,
		})
	}

	return observations, nil
}
