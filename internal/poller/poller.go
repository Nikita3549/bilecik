package poller

import (
	"context"
	"errors"
	"fmt"

	"bilecik/internal/belavia"
	"bilecik/internal/observation"
	"bilecik/internal/subscription"
)

type Deps struct {
	SubscriptionRepository *subscription.Repository
	BelaviaClient          *belavia.Client
	ObservationRepository  *observation.Repository
}

type Poller struct {
	subscriptionRepository *subscription.Repository
	belaviaClient          *belavia.Client
	observationRepository  *observation.Repository
}

func New(deps Deps) *Poller {
	return &Poller{
		subscriptionRepository: deps.SubscriptionRepository,
		belaviaClient:          deps.BelaviaClient,
		observationRepository:  deps.ObservationRepository,
	}
}

func (p *Poller) Run(ctx context.Context) error {
	targets, err := p.subscriptionRepository.GetPollerTargets(ctx)
	if err != nil {
		return fmt.Errorf("Poller Run error %v", err)
	}

	var errs []error
	for _, t := range targets {
		obs, err := p.belaviaClient.GetFromTo(ctx, t.FromIATA, t.ToIATA, t.DateFrom, t.DateTo)
		if err != nil {
			errs = append(errs, fmt.Errorf("Poller Run error, %s->%s, err %v", t.FromIATA, t.ToIATA, err))
			continue
		}

		err = p.observationRepository.CreateMany(ctx, observation.FromBelaviaAll(obs))
		if err != nil {
			errs = append(errs, fmt.Errorf("Poller Run error,  %s->%s, err %v", t.FromIATA, t.ToIATA, err))
		}
	}

	return errors.Join(errs...)
}
