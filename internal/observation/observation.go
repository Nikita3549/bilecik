package observation

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PriceObservation struct {
	ID         uuid.UUID       `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	FromIATA   string          `gorm:"column:from_iata"`
	ToIATA     string          `gorm:"column:to_iata"`
	FlightDate time.Time       `gorm:"column:flight_date"`
	Amount     decimal.Decimal `gorm:"column:amount"`
	Currency   string          `gorm:"column:currency"`
	ObservedAt time.Time       `gorm:"column:observed_at"`
	Checked    bool            `gorm:"column:checked"`
}

func (PriceObservation) TableName() string {
	return "price_observations"
}
