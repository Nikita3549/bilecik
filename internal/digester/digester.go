package digester

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bilecik/internal/bot"
	"bilecik/internal/dates"
	"bilecik/internal/money"
	"bilecik/internal/pricing"
	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var rankMarks = [...]string{"🥇", "🥈", "🥉"}

const digestHeader = "📊 Свежая сводка по подпискам"

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
	blocks := make([]string, 0, len(prices)+1)
	blocks = append(blocks, digestHeader)
	for _, p := range prices {
		blocks = append(blocks, formatSubscriptionPrices(p))
	}

	return strings.Join(blocks, "\n\n")
}

func formatSubscriptionPrices(p pricing.SubscriptionBestPrice) string {
	var b strings.Builder
	b.WriteString(subscriptionHeader(p.Subscription))

	if len(p.Tiers) == 0 {
		b.WriteString("\n\n⏳ Пока нет данных о ценах")
		return b.String()
	}

	b.WriteString("\n")
	for _, tier := range p.Tiers {
		b.WriteString("\n")
		b.WriteString(formatTier(tier))
	}

	return b.String()
}

func subscriptionHeader(s subscription.Subscription) string {
	period := dates.FormatRange(s.DateFrom, s.DateTo)
	if s.DateFrom.Year() == s.DateTo.Year() {
		period += fmt.Sprintf(" %d", s.DateTo.Year())
	}
	return fmt.Sprintf("✈️ %s → %s\n- %s", s.FromLabel(), s.ToLabel(), period)
}

func formatTier(t pricing.PriceTier) string {
	mark := fmt.Sprintf("%d)", t.Rank)
	if t.Rank >= 1 && t.Rank <= len(rankMarks) {
		mark = rankMarks[t.Rank-1]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s %s:", mark, money.FormatWhole(t.Amount), t.Currency)
	for _, r := range t.Dates {
		b.WriteString("\n• ")
		b.WriteString(dates.FormatRange(r.From, r.To))
	}

	return b.String()
}
