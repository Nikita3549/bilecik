package observation

import (
	"context"

	db "bilecik/pkg"
)

type Repository struct {
	*db.DB
}

func NewRepository(db *db.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

func (repo *Repository) CreateMany(ctx context.Context, obs []PriceObservation) error {
	if len(obs) == 0 {
		return nil
	}
	return repo.DB.WithContext(ctx).Create(obs).Error
}
