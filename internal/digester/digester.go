package digester

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bilecik/internal/bot"
	"bilecik/internal/dates"
	"bilecik/internal/pricing"
	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var rankMarks = [...]string{"🥇", "🥈", "🥉"}

type Digester struct {
	pricingRepository *pricing.Repository
	api               *tgbotapi.BotAPI
}

func NewDigester(pricingRepository *pricing.Repository, api *tgbotapi.BotAPI) *Digester {
	return &Digester{
		pricingRepository,
		api,
	}
}

const TopPricesAmount = 3

func (d *Digester) Run(ctx context.Context) {
	bestPrices, err := d.pricingRepository.ListWithBestPrices(ctx, TopPricesAmount)
	if err != nil {
		log.Printf("Digest error: %s", err)
		return
	}

	groupedBestPrices := make(map[int64][]pricing.SubscriptionBestPrice)
	for _, price := range bestPrices {
		groupedBestPrices[price.Subscription.TelegramID] = append(groupedBestPrices[price.Subscription.TelegramID], price)
	}

	for telegramID, prices := range groupedBestPrices {
		d.SendDailyDigest(telegramID, prices)
	}
}

func (d *Digester) SendDailyDigest(telegramID int64, prices []pricing.SubscriptionBestPrice) {
	bot.Send(d.api, telegramID, formatDailyDigest(prices))
}

func formatDailyDigest(prices []pricing.SubscriptionBestPrice) string {
	var b strings.Builder
	b.WriteString("📊 Свежая сводка по твоим подпискам\n")
	for _, p := range prices {
		b.WriteString("\n")
		b.WriteString(subscriptionHeader(p.Subscription))
		b.WriteString("\n")

		if len(p.Tiers) == 0 {
			b.WriteString("⏳ пока нет данных о ценах\n")
			continue
		}

		for _, tier := range p.Tiers {
			b.WriteString(formatTier(tier))
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func subscriptionHeader(s subscription.Subscription) string {
	period := dates.FormatRange(s.DateFrom, s.DateTo)
	if s.DateFrom.Year() == s.DateTo.Year() {
		period += fmt.Sprintf(" %d", s.DateTo.Year())
	}
	return fmt.Sprintf("✈️ %s → %s · %s", s.FromLabel(), s.ToLabel(), period)
}

func formatTier(t pricing.PriceTier) string {
	mark := fmt.Sprintf("%d)", t.Rank)
	if t.Rank >= 1 && t.Rank <= len(rankMarks) {
		mark = rankMarks[t.Rank-1]
	}

	ranges := make([]string, len(t.Dates))
	for i, r := range t.Dates {
		ranges[i] = dates.FormatRange(r.From, r.To)
	}

	return fmt.Sprintf("%s %s %s — %s", mark, t.Amount.String(), t.Currency, strings.Join(ranges, " · "))
}
