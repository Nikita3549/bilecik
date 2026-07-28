package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bilecik/internal/airport"
	"bilecik/internal/belavia"
	"bilecik/internal/bot"
	"bilecik/internal/configs"
	"bilecik/internal/observation"
	"bilecik/internal/poller"
	"bilecik/internal/scheduler"
	"bilecik/internal/subscription"
	db "bilecik/pkg"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
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

	var esClient *elasticsearch.Client
	if conf.ElasticConfig.URL != "" {
		var err error
		esClient, err = airport.NewElasticClient(conf.ElasticConfig.URL)
		if err != nil {
			log.Fatalf("elastic init: %v", err)
		}
	}
	airportRepository := airport.NewRepository(database, esClient)
	belaviaClient := belavia.NewClient()

	p := poller.New(poller.Deps{
		SubscriptionRepository: subscriptionRepository,
		BelaviaClient:          belaviaClient,
		ObservationRepository:  observationRepository,
	})

	g, gctx := errgroup.WithContext(ctx)

	if esClient != nil {
		g.Go(func() error {
			// Search falls back to Postgres until the index is ready, so a
			// failed sync degrades quality instead of killing the bot.
			if err := airport.SyncElastic(gctx, esClient, database); err != nil {
				log.Printf("elastic sync failed: %v", err)
			}
			return nil
		})
	}

	g.Go(func() error {
		return bot.Run(gctx, conf.TgBotConfig.Token, subscriptionRepository, airportRepository)
	})

	g.Go(func() error {
		scheduler.New().RunEvery(gctx, 30*time.Minute, 5*time.Minute, p)
		return nil
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		log.Fatalf("shutdown with error: %v", err)
	}
}
