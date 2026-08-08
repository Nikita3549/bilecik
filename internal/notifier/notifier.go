package notifier

import (
	"context"
	"fmt"
	"strings"

	"bilecik/internal/bot"
	"bilecik/internal/dates"
	"bilecik/internal/observation"
	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// maxMessageLen keeps a single alert under Telegram's 4096-character cap. The
// budget is counted in bytes, which for Cyrillic and emoji is never smaller
// than the UTF-16 units Telegram actually counts, so the margin is safe.
const maxMessageLen = 3500

const alertHeader = "🔻 Цена упала ниже твоего порога!"

type Notifier struct {
	api                   *tgbotapi.BotAPI
	observationRepository *observation.Repository
}

func NewNotifier(api *tgbotapi.BotAPI, observationRepository *observation.Repository) *Notifier {
	return &Notifier{
		api:                   api,
		observationRepository: observationRepository,
	}
}

func (n *Notifier) Run(ctx context.Context) error {
	hits, err := n.observationRepository.ListBelowThreshold(ctx)
	if err != nil {
		return fmt.Errorf("list below threshold: %w", err)
	}

	hitsByTelegramID := make(map[int64][]observation.ThresholdHit)

	for _, hit := range hits {
		hitsByTelegramID[hit.TelegramID] = append(hitsByTelegramID[hit.TelegramID], hit)
	}

	for telegramID, userHits := range hitsByTelegramID {
		if err := ctx.Err(); err != nil {
			return err
		}
		n.SendThresholdAlert(telegramID, userHits)
	}

	return nil
}

func (n *Notifier) SendThresholdAlert(telegramID int64, hits []observation.ThresholdHit) {
	for _, message := range formatThresholdAlert(hits) {
		bot.Send(n.api, telegramID, message)
	}
}

// formatThresholdAlert splits the hits across as many messages as it takes to
// stay under maxMessageLen: a user with a large backlog would otherwise produce
// one oversized message that Telegram rejects wholesale.
func formatThresholdAlert(hits []observation.ThresholdHit) []string {
	if len(hits) == 0 {
		return nil
	}

	var (
		messages []string
		b        strings.Builder
	)
	b.WriteString(alertHeader)

	for _, h := range hits {
		line := fmt.Sprintf("\n✈️ %s → %s · %s — %s %s (порог %s)",
			subscription.PlaceLabel(h.FromCity, h.FromIATA),
			subscription.PlaceLabel(h.ToCity, h.ToIATA),
			dates.FormatDay(h.FlightDate),
			h.Amount.String(), h.Currency, h.Threshold.String())

		if b.Len()+len(line) > maxMessageLen {
			messages = append(messages, b.String())
			b.Reset()
			b.WriteString(alertHeader)
		}
		b.WriteString(line)
	}

	return append(messages, b.String())
}
