package belavia

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Belavia API constraints. DO NOT CHANGE + read docs
const apiDaysCount = 7

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetFromTo(from, to string, startDate, endDate time.Time) ([]Observation, error) {
	totalDays := int(endDate.Sub(startDate).Hours() / 24)
	if totalDays <= 0 {
		return nil, nil
	}

	var observations []Observation
	for offset := 0; offset < totalDays; offset += apiDaysCount {
		chunkStart := startDate.AddDate(0, 0, offset)
		chunk, err := fetchWindow(from, to, chunkStart)
		if err != nil {
			return nil, err
		}
		observations = append(observations, chunk...)
	}

	trimmed := observations[:0]
	for _, obs := range observations {
		if obs.FlightDate.Before(startDate) || !obs.FlightDate.Before(endDate) {
			continue
		}
		trimmed = append(trimmed, obs)
	}

	return trimmed, nil
}

func fetchWindow(from, to string, startDate time.Time) ([]Observation, error) {
	data := buildOneWayRequest(from, to, startDate)
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://webapi.belavia.by/graphql/query/nemo", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	apiResponse, err := decodeBelaviaResponse(body)
	if err != nil {
		return nil, err
	}

	return parseObservations(apiResponse, from, to, time.Now())
}
