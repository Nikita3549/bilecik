package bot

import (
	"context"
	"log"
	"strings"
	"time"

	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

const helpText = `Я слежу за ценами на рейсы Belavia и пишу, когда дешевеет.

Команды:
/subscribe FROM TO ДАТА_ОТ ДАТА_ДО [ЦЕНА] — подписаться на маршрут
/list — мои подписки
/unsubscribe ID — удалить подписку
/help — эта справка

Пример:
/subscribe MSQ IST 2026-08-01 2026-08-10 250

FROM и TO — IATA-коды аэропортов (3 латинские буквы).
Даты — в формате ГГГГ-ММ-ДД. ЦЕНА (в BYN) необязательна: с ней я напишу
только когда билет станет дешевле указанной суммы.`

type Handlers struct {
	subs *subscription.Repository
}

func NewHandlers(subs *subscription.Repository) *Handlers {
	return &Handlers{subs: subs}
}

func (h *Handlers) Start(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	send(api, msg.Chat.ID, "Привет! Я bilecik — ловлю дешёвые билеты Belavia.\n\n"+helpText)
}

func (h *Handlers) Help(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	send(api, msg.Chat.ID, helpText)
}

func (h *Handlers) Subscribe(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	parsed, err := parseSubscribeArgs(msg.CommandArguments(), time.Now())
	if err != nil {
		send(api, msg.Chat.ID, "❌ "+err.Error()+
			"\n\nПример: /subscribe MSQ IST 2026-08-01 2026-08-10 250")
		return
	}

	sub := &subscription.Subscription{
		TelegramID: msg.Chat.ID,
		FromIATA:   parsed.FromIATA,
		ToIATA:     parsed.ToIATA,
		DateFrom:   parsed.DateFrom,
		DateTo:     parsed.DateTo,
		Threshold:  parsed.Threshold,
	}
	if err := h.subs.Create(ctx, sub); err != nil {
		log.Printf("subscribe: create failed: %v", err)
		send(api, msg.Chat.ID, "Не смог сохранить подписку, попробуй ещё раз позже.")
		return
	}

	send(api, msg.Chat.ID, "✅ Подписка создана:\n"+formatSubscription(*sub))
}

func (h *Handlers) List(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	subs, err := h.subs.ListByTelegramID(ctx, msg.Chat.ID)
	if err != nil {
		log.Printf("list: query failed: %v", err)
		send(api, msg.Chat.ID, "Не смог достать подписки, попробуй позже.")
		return
	}
	if len(subs) == 0 {
		send(api, msg.Chat.ID, "У тебя пока нет подписок. Создай первую через /subscribe.")
		return
	}

	var b strings.Builder
	b.WriteString("Твои подписки:\n\n")
	for _, s := range subs {
		b.WriteString(formatSubscription(s))
		b.WriteString("\n\n")
	}
	b.WriteString("Удалить: /unsubscribe ID")
	send(api, msg.Chat.ID, b.String())
}

func (h *Handlers) Unsubscribe(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	arg := strings.TrimSpace(msg.CommandArguments())
	id, err := uuid.Parse(arg)
	if err != nil {
		send(api, msg.Chat.ID, "Укажи ID подписки: /unsubscribe ID\nПосмотреть ID — /list")
		return
	}

	deleted, err := h.subs.Delete(ctx, id, msg.Chat.ID)
	if err != nil {
		log.Printf("unsubscribe: delete failed: %v", err)
		send(api, msg.Chat.ID, "Не смог удалить, попробуй позже.")
		return
	}
	if !deleted {
		send(api, msg.Chat.ID, "Подписка не найдена.")
		return
	}
	send(api, msg.Chat.ID, "🗑 Подписка удалена.")
}

func (h *Handlers) Fallback(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.IsCommand() {
		send(api, msg.Chat.ID, "Не знаю такой команды. Список — /help")
		return
	}
	send(api, msg.Chat.ID, "Я работаю по командам. Начни с /help")
}

func formatSubscription(s subscription.Subscription) string {
	line := s.FromIATA + " → " + s.ToIATA + "  " +
		s.DateFrom.Format(dateLayout) + " … " + s.DateTo.Format(dateLayout)
	if s.Threshold.Valid {
		line += "\nцель: до " + s.Threshold.Decimal.String() + " BYN"
	}
	line += "\nID: " + s.ID.String()
	return line
}

func send(api *tgbotapi.BotAPI, chatID int64, text string) {
	if _, err := api.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		log.Printf("send to %d failed: %v", chatID, err)
	}
}
