package bot

import (
	"testing"
	"time"

	"bilecik/internal/subscription"

	"github.com/shopspring/decimal"
)

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func sub(threshold decimal.NullDecimal) subscription.Subscription {
	return subscription.Subscription{
		FromIATA: "MSQ", FromCity: "Минск",
		ToIATA: "DXB", ToCity: "Дубай",
		DateFrom:  day("2027-01-25"),
		DateTo:    day("2027-01-29"),
		Threshold: threshold,
	}
}

func threshold(s string) decimal.NullDecimal {
	return decimal.NullDecimal{Decimal: decimal.RequireFromString(s), Valid: true}
}

func TestFormatSubscriptionItem(t *testing.T) {
	t.Run("with threshold", func(t *testing.T) {
		want := "✈️ 1. Минск (MSQ) → Дубай (DXB)\n📅 25.01.2027 — 29.01.2027 | 🎯 до 900 BYN"
		if got := formatSubscriptionItem(1, sub(threshold("900.00"))); got != want {
			t.Errorf("formatSubscriptionItem =\n%q\nwant\n%q", got, want)
		}
	})

	t.Run("without threshold", func(t *testing.T) {
		want := "✈️ 2. Минск (MSQ) → Дубай (DXB)\n📅 25.01.2027 — 29.01.2027 | 🎯 без порога"
		if got := formatSubscriptionItem(2, sub(decimal.NullDecimal{})); got != want {
			t.Errorf("formatSubscriptionItem =\n%q\nwant\n%q", got, want)
		}
	})
}

func TestFormatSubscriptionCreated(t *testing.T) {
	want := "✅ Подписка успешно оформлена!\n\n" +
		"✈️ Маршрут: Минск (MSQ) → Дубай (DXB)\n" +
		"📅 Даты: 25.01.2027 — 29.01.2027\n" +
		"🎯 Цель: до 900 BYN\n\n" +
		"Я пришлю уведомление, как только цена упадёт до нужного уровня."

	if got := formatSubscriptionCreated(sub(threshold("900.00"))); got != want {
		t.Errorf("formatSubscriptionCreated =\n%q\nwant\n%q", got, want)
	}
}
