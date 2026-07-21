package subscription

import (
	"context"

	db "bilecik/pkg"
)

type Repository struct {
	db *db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) Create(ctx context.Context, subscription *Subscription) error {
	return repo.db.WithContext(ctx).Create(subscription).Error
}

func (repo *Repository) GetPollerTargets(ctx context.Context) ([]PollerTarget, error) {
	var pollerTargets []PollerTarget
	err := repo.db.WithContext(ctx).Raw(`
		SELECT 
			from_iata, 
			to_iata, 
			MIN(date_from) as date_from, 
			MAX(date_to) as date_to
		FROM subscriptions 
		GROUP BY to_iata, from_iata
	`).Scan(&pollerTargets).Error

	return pollerTargets, err
}
