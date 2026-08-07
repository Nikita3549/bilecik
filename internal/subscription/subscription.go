package subscription

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Subscription struct {
	ID         uuid.UUID           `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TelegramID int64               `gorm:"column:telegram_id"`
	FromIATA   string              `gorm:"column:from_iata"`
	ToIATA     string              `gorm:"column:to_iata"`
	FromCity   string              `gorm:"column:from_city"`
	ToCity     string              `gorm:"column:to_city"`
	DateFrom   time.Time           `gorm:"column:date_from"`
	DateTo     time.Time           `gorm:"column:date_to"`
	Threshold  decimal.NullDecimal `gorm:"column:threshold"`
	CreatedAt  time.Time           `gorm:"column:created_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

// FromLabel returns «Минск (MSQ)», or the bare IATA code when the city is
// unknown (e.g. the code was typed by hand and is missing from the catalog).
func (s Subscription) FromLabel() string {
	return placeLabel(s.FromCity, s.FromIATA)
}

func (s Subscription) ToLabel() string {
	return placeLabel(s.ToCity, s.ToIATA)
}

func placeLabel(city, iata string) string {
	if city == "" {
		return iata
	}
	return city + " (" + iata + ")"
}

type PollerTarget struct {
	FromIATA string    `gorm:"column:from_iata"`
	ToIATA   string    `gorm:"column:to_iata"`
	DateFrom time.Time `gorm:"column:date_from"`
	DateTo   time.Time `gorm:"column:date_to"`
}
