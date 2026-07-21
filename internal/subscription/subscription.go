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
	DateFrom   time.Time           `gorm:"column:date_from"`
	DateTo     time.Time           `gorm:"column:date_to"`
	Threshold  decimal.NullDecimal `gorm:"column:threshold"`
	CreatedAt  time.Time           `gorm:"column:created_at"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type PollerTarget struct {
	FromIATA string    `gorm:"column:from_iata"`
	ToIATA   string    `gorm:"column:to_iata"`
	DateFrom time.Time `gorm:"column:date_from"`
	DateTo   time.Time `gorm:"column:date_to"`
}
