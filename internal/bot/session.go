package bot

import (
	"context"
	"log"
	"sync"
	"time"

	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/shopspring/decimal"
)

type step int

const (
	stepFrom step = iota
	stepTo
	stepDateFrom
	stepDateTo
	stepThreshold
)

type subscribeSession struct {
	step      step
	fromIATA  string
	toIATA    string
	dateFrom  time.Time
	dateTo    time.Time
	threshold decimal.NullDecimal
}

type sessions struct {
	mu sync.Mutex
	m  map[int64]*subscribeSession
}

func newSessions() *sessions {
	return &sessions{m: make(map[int64]*subscribeSession)}
}

func (s *sessions) get(chatID int64) (*subscribeSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.m[chatID]
	return sess, ok
}

func (s *sessions) start(chatID int64) *subscribeSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := &subscribeSession{step: stepFrom}
	s.m[chatID] = sess
	return sess
}

func (s *sessions) delete(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, chatID)
}

func (h *Handlers) intercept(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) bool {
	if msg.Command() == "cancel" {
		return false
	}
	sess, ok := h.sessions.get(msg.Chat.ID)
	if !ok {
		return false
	}
	h.handleStep(ctx, api, msg, sess)
	return true
}

func (h *Handlers) handleStep(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message, sess *subscribeSession) {
	chatID := msg.Chat.ID
	text := msg.Text

	switch sess.step {
	case stepFrom:
		iata, err := validateIATA(text)
		if err != nil {
			send(api, chatID, "❌ "+err.Error())
			return
		}
		sess.fromIATA = iata
		sess.step = stepTo
		send(api, chatID, "Куда летим? Пришли IATA-код города назначения (например IST).")

	case stepTo:
		iata, err := validateIATA(text)
		if err != nil {
			send(api, chatID, "❌ "+err.Error())
			return
		}
		if iata == sess.fromIATA {
			send(api, chatID, "❌ Город назначения совпадает с городом вылета. Пришли другой.")
			return
		}
		sess.toIATA = iata
		sess.step = stepDateFrom
		send(api, chatID, "С какой даты искать? Формат ГГГГ-ММ-ДД (например 2026-08-01).")

	case stepDateFrom:
		date, err := parseFlightDate(text)
		if err != nil {
			send(api, chatID, "❌ "+err.Error())
			return
		}
		if date.Before(startOfDay(time.Now())) {
			send(api, chatID, "❌ Эта дата уже в прошлом. Пришли другую.")
			return
		}
		sess.dateFrom = date
		sess.step = stepDateTo
		send(api, chatID, "По какую дату? Формат ГГГГ-ММ-ДД.")

	case stepDateTo:
		date, err := parseFlightDate(text)
		if err != nil {
			send(api, chatID, "❌ "+err.Error())
			return
		}
		if date.Before(sess.dateFrom) {
			send(api, chatID, "❌ Дата «по» раньше даты «с». Пришли другую.")
			return
		}
		sess.dateTo = date
		sess.step = stepThreshold
		send(api, chatID, "Целевая цена в BYN? Пришли число (например 250) или «-», чтобы следить без порога.")

	case stepThreshold:
		threshold, err := parseThreshold(text)
		if err != nil {
			send(api, chatID, "❌ "+err.Error())
			return
		}
		sess.threshold = threshold

		sub := &subscription.Subscription{
			TelegramID: chatID,
			FromIATA:   sess.fromIATA,
			ToIATA:     sess.toIATA,
			DateFrom:   sess.dateFrom,
			DateTo:     sess.dateTo,
			Threshold:  sess.threshold,
		}
		h.sessions.delete(chatID)

		if err := h.subs.Create(ctx, sub); err != nil {
			log.Printf("subscribe: create failed: %v", err)
			send(api, chatID, "Не смог сохранить подписку, попробуй позже: /subscribe")
			return
		}
		send(api, chatID, "✅ Подписка создана:\n"+formatSubscription(*sub))
	}
}
