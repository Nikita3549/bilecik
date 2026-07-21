package main

import (
	"fmt"
	"log"
	"time"

	"bilecik/internal/configs"
	"bilecik/internal/observation"
	"bilecik/internal/subscription"
	db "bilecik/pkg"

	"github.com/shopspring/decimal"
)

func main() {
	conf := configs.LoadConfig()

	database := db.NewDB(conf)
	defer database.Close()

	subRepo := subscription.NewSubscriptionRepository(database)
	obsRepo := observation.NewObservationRepository(database)

	sub := &subscription.Subscription{
		TelegramID: 123456789,
		FromIATA:   "MSQ",
		ToIATA:     "IST",
		DateFrom:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		Threshold:  decimal.NullDecimal{Decimal: decimal.RequireFromString("150.00"), Valid: true},
	}
	if err := subRepo.Create(sub); err != nil {
		log.Fatalf("create subscription: %s", err)
	}
	fmt.Printf("subscription created: id=%s\n", sub.ID)

	obs := &observation.PriceObservation{
		FromIATA:   "MSQ",
		ToIATA:     "IST",
		FlightDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Amount:     decimal.RequireFromString("142.50"),
		Currency:   "EUR",
		ObservedAt: time.Now(),
	}
	if err := obsRepo.Create(obs); err != nil {
		log.Fatalf("create observation: %s", err)
	}
	fmt.Printf("observation created: id=%s\n", obs.ID)
}
