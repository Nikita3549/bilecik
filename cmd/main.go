package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bilecik/internal/airport"
	"bilecik/internal/belavia"
	"bilecik/internal/bot"
	"bilecik/internal/configs"
	"bilecik/internal/digester"
	"bilecik/internal/notifier"
	"bilecik/internal/observation"
	"bilecik/internal/poller"
	"bilecik/internal/pricing"
	"bilecik/internal/scheduler"
	"bilecik/internal/subscription"
	db "bilecik/pkg"

	_ "time/tzdata"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf := configs.LoadConfig()

	api, err := tgbotapi.NewBotAPI(conf.TgBotConfig.Token)
	if err != nil {
		log.Fatalf("bot init: %v", err)
	}

	database := db.NewDB(conf)
	defer database.Close()

	subscriptionRepository := subscription.NewRepository(database)
	observationRepository := observation.NewRepository(database)
	pricingRepository := pricing.NewRepository(database)

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
	n := notifier.NewNotifier(api, observationRepository)
	d := digester.NewDigester(pricingRepository, api)

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
		return bot.Run(gctx, api, subscriptionRepository, airportRepository)
	})

	g.Go(func() error {
		scheduler.New().RunEvery(gctx, 5*time.Minute, 1*time.Minute, n)
		return nil
	})

	g.Go(func() error {
		scheduler.New().RunEvery(gctx, 30*time.Minute, 5*time.Minute, p)
		return nil
	})

	g.Go(func() error {
		loc, err := time.LoadLocation("Europe/Minsk")
		if err != nil {
			log.Printf("load location failed, falling back to UTC: %v", err)
			loc = time.UTC
		}
		c := cron.New(cron.WithLocation(loc))

		if _, err := c.AddFunc("0 9 * * *", func() {
			d.Run(gctx)
		}); err != nil {
			return fmt.Errorf("cron add: %w", err)
		}

		c.Start()
		<-gctx.Done()
		<-c.Stop().Done()
		return gctx.Err()
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		log.Fatalf("shutdown with error: %v", err)
	}
}
