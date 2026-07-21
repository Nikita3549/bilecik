package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bilecik/internal/belavia"
	"bilecik/internal/configs"
	"bilecik/internal/observation"
	"bilecik/internal/poller"
	"bilecik/internal/scheduler"
	"bilecik/internal/subscription"
	db "bilecik/pkg"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := scheduler.New()
	conf := configs.LoadConfig()
	db := db.NewDB(conf)

	subscriptionRepository := subscription.NewRepository(db)
	belaviaClient := belavia.NewClient()
	observationRepository := observation.NewRepository(db)

	p := poller.New(poller.Deps{
		SubscriptionRepository: subscriptionRepository,
		BelaviaClient:          belaviaClient,
		ObservationRepository:  observationRepository,
	})

	s.RunEvery(ctx, 30*time.Minute, 5*time.Minute, p)
}
