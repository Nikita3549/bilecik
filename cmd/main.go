package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bilecik/internal/belavia"
	"bilecik/internal/bot"
	"bilecik/internal/configs"
	"bilecik/internal/observation"
	"bilecik/internal/poller"
	"bilecik/internal/scheduler"
	"bilecik/internal/subscription"
	db "bilecik/pkg"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf := configs.LoadConfig()
	database := db.NewDB(conf)
	defer database.Close()

	subscriptionRepository := subscription.NewRepository(database)
	observationRepository := observation.NewRepository(database)
	belaviaClient := belavia.NewClient()

	p := poller.New(poller.Deps{
		SubscriptionRepository: subscriptionRepository,
		BelaviaClient:          belaviaClient,
		ObservationRepository:  observationRepository,
	})

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return bot.Run(gctx, conf.TgBotConfig.Token, subscriptionRepository)
	})

	g.Go(func() error {
		scheduler.New().RunEvery(gctx, 30*time.Minute, 5*time.Minute, p)
		return nil
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		log.Fatalf("shutdown with error: %v", err)
	}
}
