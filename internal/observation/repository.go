package observation

import db "bilecik/pkg"

type ObservationRepository struct {
	*db.DB
}

func NewObservationRepository(db *db.DB) *ObservationRepository {
	return &ObservationRepository{
		DB: db,
	}
}

func (repo *ObservationRepository) Create(priceObservation *PriceObservation) error {
	return repo.DB.Create(priceObservation).Error
}
