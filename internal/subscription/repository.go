package subscription

import (
	"context"

	db "bilecik/pkg"

	"github.com/google/uuid"
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

func (repo *Repository) ListByTelegramID(ctx context.Context, telegramID int64) ([]Subscription, error) {
	var subs []Subscription
	err := repo.db.WithContext(ctx).
		Where("telegram_id = ?", telegramID).
		Order("created_at DESC").
		Find(&subs).Error
	return subs, err
}

func (repo *Repository) Delete(ctx context.Context, id uuid.UUID, telegramID int64) (bool, error) {
	res := repo.db.WithContext(ctx).
		Where("id = ? AND telegram_id = ?", id, telegramID).
		Delete(&Subscription{})
	return res.RowsAffected > 0, res.Error
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
