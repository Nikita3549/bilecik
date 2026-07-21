package subscription

import db "bilecik/pkg"

type SubscriptionRepository struct {
	db *db.DB
}

func NewSubscriptionRepository(db *db.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (repo *SubscriptionRepository) Create(subscription *Subscription) error{
	return repo.db.DB.Create(subscription).Error
}
