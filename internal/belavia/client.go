package belavia

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand/v2"
	"net/http"
	"time"
)

// Belavia API constraints. DO NOT CHANGE + read docs
const apiDaysCount = 7

const (
	defaultMinPause = 2 * time.Second
	defaultMaxPause = 8 * time.Second
)

type Client struct {
	minPause time.Duration
	maxPause time.Duration
}

func NewClient() *Client {
	return &Client{
		minPause: defaultMinPause,
		maxPause: defaultMaxPause,
	}
}

func (c *Client) GetFromTo(ctx context.Context, from, to string, startDate, endDate time.Time) ([]Observation, error) {
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if totalDays <= 0 {
		return nil, nil
	}

	var observations []Observation
	for offset := 0; offset < totalDays; offset += apiDaysCount {
		if err := c.pause(ctx); err != nil {
			return nil, err
		}

		chunkStart := startDate.AddDate(0, 0, offset)
		chunk, err := fetchWindow(ctx, from, to, chunkStart)
		if err != nil {
			return nil, err
		}
		observations = append(observations, chunk...)
	}

	trimmed := observations[:0]
	for _, obs := range observations {
		if obs.FlightDate.Before(startDate) || obs.FlightDate.After(endDate) {
			continue
		}
		trimmed = append(trimmed, obs)
	}

	return trimmed, nil
}

func (c *Client) pause(ctx context.Context) error {
	d := c.minPause
	if c.maxPause > c.minPause {
		d += rand.N(c.maxPause - c.minPause)
	}
	if d <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func fetchWindow(ctx context.Context, from, to string, startDate time.Time) ([]Observation, error) {
	data := buildOneWayRequest(from, to, startDate)
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://webapi.belavia.by/graphql/query/nemo", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
