package bot

import (
	"context"
	"log"
	"strings"

	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

const helpText = `Я слежу за ценами на рейсы Belavia и пишу, когда дешевеет.

Команды:
/subscribe — подписаться на маршрут (проведу по шагам)
/list — мои подписки
/unsubscribe ID — удалить подписку
/cancel — прервать текущий диалог
/help — эта справка`

type Handlers struct {
	subs     *subscription.Repository
	sessions *sessions
}

func NewHandlers(subs *subscription.Repository) *Handlers {
	return &Handlers{
		subs:     subs,
		sessions: newSessions(),
	}
}

func (h *Handlers) Interceptor() Interceptor {
	return h.intercept
}

func (h *Handlers) Start(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	send(api, msg.Chat.ID, "Привет! Я bilecik — ловлю дешёвые билеты Belavia.\n\n"+helpText)
}

func (h *Handlers) Help(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	send(api, msg.Chat.ID, helpText)
}

func (h *Handlers) Subscribe(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	h.sessions.start(msg.Chat.ID)
	send(api, msg.Chat.ID, "Откуда летим? Пришли IATA-код города вылета (например MSQ).\n\nВ любой момент — /cancel.")
}

func (h *Handlers) Cancel(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if _, ok := h.sessions.get(msg.Chat.ID); !ok {
		send(api, msg.Chat.ID, "Нечего отменять.")
		return
	}
	h.sessions.delete(msg.Chat.ID)
	send(api, msg.Chat.ID, "Отменил. Начать заново — /subscribe.")
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
	b.WriteString("Удалить — /unsubscribe")
	send(api, msg.Chat.ID, b.String())
}

const unsubPrefix = "unsub"

func (h *Handlers) Unsubscribe(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	subs, err := h.subs.ListByTelegramID(ctx, msg.Chat.ID)
	if err != nil {
		log.Printf("unsubscribe: query failed: %v", err)
		send(api, msg.Chat.ID, "Не смог достать подписки, попробуй позже.")
		return
	}
	if len(subs) == 0 {
		send(api, msg.Chat.ID, "У тебя пока нет подписок. Создай первую через /subscribe.")
		return
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(subs))
	for _, s := range subs {
		btn := tgbotapi.NewInlineKeyboardButtonData(subscriptionLabel(s), unsubPrefix+":"+s.ID.String())
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	out := tgbotapi.NewMessage(msg.Chat.ID, "Какую подписку удалить?")
	out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := api.Send(out); err != nil {
		log.Printf("unsubscribe: send picker failed: %v", err)
	}
}

func (h *Handlers) UnsubscribeCallback(ctx context.Context, api *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	_, idStr, _ := strings.Cut(cq.Data, ":")
	id, err := uuid.Parse(idStr)
	if err != nil {
		answerCallback(api, cq.ID, "Не понял, какую подписку удалить.")
		return
	}

	deleted, err := h.subs.Delete(ctx, id, cq.Message.Chat.ID)
	if err != nil {
		log.Printf("unsubscribe: delete failed: %v", err)
		answerCallback(api, cq.ID, "Не смог удалить, попробуй позже.")
		return
	}
	if !deleted {
		answerCallback(api, cq.ID, "Подписка не найдена.")
		return
	}

	answerCallback(api, cq.ID, "Удалено")
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, "🗑 Подписка удалена.")
	if _, err := api.Send(edit); err != nil {
		log.Printf("unsubscribe: edit message failed: %v", err)
	}
}

func (h *Handlers) Fallback(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.IsCommand() {
		send(api, msg.Chat.ID, "Не знаю такой команды. Список — /help")
		return
	}
	send(api, msg.Chat.ID, "Я работаю по командам. Начни с /help")
}

func subscriptionLabel(s subscription.Subscription) string {
	return s.FromIATA + " → " + s.ToIATA + "  " +
		s.DateFrom.Format(dateLayout) + " … " + s.DateTo.Format(dateLayout)
}

func formatSubscription(s subscription.Subscription) string {
	line := subscriptionLabel(s)
	if s.Threshold.Valid {
		line += "\nцель: до " + s.Threshold.Decimal.String() + " BYN"
	}
	return line
}

func send(api *tgbotapi.BotAPI, chatID int64, text string) {
	if _, err := api.Send(tgbotapi.NewMessage(chatID, text)); err != nil {
		log.Printf("send to %d failed: %v", chatID, err)
	}
}

func answerCallback(api *tgbotapi.BotAPI, callbackID, text string) {
	if _, err := api.Request(tgbotapi.NewCallback(callbackID, text)); err != nil {
		log.Printf("answer callback failed: %v", err)
	}
}
