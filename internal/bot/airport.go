package bot

import (
	"context"
	"log"
	"regexp"
	"strings"

	"bilecik/internal/airport"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const airportPrefix = "ap"

var iataRe = regexp.MustCompile(`^[A-Za-z]{3}$`)

func (s step) airportField() string {
	switch s {
	case stepFrom:
		return "from"
	case stepTo:
		return "to"
	}
	return ""
}

func (h *Handlers) handleAirportSearch(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message, sess *subscribeSession) {
	chatID := msg.Chat.ID
	query := strings.TrimSpace(msg.Text)
	if query == "" {
		Send(api, chatID, "Напиши город или аэропорт текстом, например «Минск» или MSQ.")
		return
	}

	airports, err := h.airports.Search(ctx, query)
	if err != nil {
		log.Printf("airport search %q failed: %v", query, err)
		Send(api, chatID, "Поиск сейчас недоступен, попробуй ещё раз чуть позже.")
		return
	}

	switch len(airports) {
	case 0:
		// Unknown to the catalog, but a bare IATA code is accepted as-is:
		// the catalog is curated and must not block real routes.
		if iataRe.MatchString(query) {
			code := strings.ToUpper(query)
			h.removePicker(api, chatID, sess)
			confirm, ok := h.applyAirport(sess, code, "", code)
			if !ok {
				Send(api, chatID, "❌ Город назначения совпадает с городом вылета. Пришли другой.")
				return
			}
			Send(api, chatID, confirm+"\n\n"+nextQuestion(sess.step))
			return
		}
		Send(api, chatID, "Ничего не нашёл. Попробуй иначе — другое написание, или пришли IATA-код (3 латинские буквы).")
	case 1:
		h.removePicker(api, chatID, sess)
		confirm, ok := h.applyAirport(sess, airports[0].IATACode, airports[0].Place(), airports[0].Label())
		if !ok {
			Send(api, chatID, "❌ Город назначения совпадает с городом вылета. Пришли другой.")
			return
		}
		Send(api, chatID, confirm+"\n\n"+nextQuestion(sess.step))
	default:
		h.removePicker(api, chatID, sess)
		field := sess.step.airportField()
		labels := pickerLabels(airports)
		rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(airports))
		for i, a := range airports {
			btn := tgbotapi.NewInlineKeyboardButtonData(labels[i], airportPrefix+":"+field+":"+a.IATACode)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		}
		out := tgbotapi.NewMessage(chatID, "Нашёл несколько вариантов:")
		out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		sent, err := api.Send(out)
		if err != nil {
			log.Printf("send airport picker failed: %v", err)
			return
		}
		sess.pickerMsgID = sent.MessageID
	}
}

func (h *Handlers) AirportCallback(ctx context.Context, api *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	parts := strings.SplitN(cq.Data, ":", 3)
	if len(parts) != 3 || cq.Message == nil {
		answerCallback(api, cq.ID, "Это устаревшая кнопка.")
		return
	}
	field, code := parts[1], parts[2]
	chatID := cq.Message.Chat.ID

	sess, ok := h.sessions.get(chatID)
	if !ok || sess.step.airportField() != field {
		answerCallback(api, cq.ID, "Это устаревшая кнопка.")
		return
	}

	ap, err := h.airports.GetByIATA(ctx, code)
	if err != nil {
		log.Printf("airport by iata %q failed: %v", code, err)
		answerCallback(api, cq.ID, "Не получилось, попробуй ещё раз.")
		return
	}
	if ap == nil {
		answerCallback(api, cq.ID, "Это устаревшая кнопка.")
		return
	}

	confirm, ok := h.applyAirport(sess, ap.IATACode, ap.Place(), ap.Label())
	if !ok {
		answerCallback(api, cq.ID, "Совпадает с городом вылета, выбери другой.")
		return
	}

	answerCallback(api, cq.ID, "")
	edit := tgbotapi.NewEditMessageText(chatID, cq.Message.MessageID, confirm)
	if _, err := api.Send(edit); err != nil {
		log.Printf("edit airport picker failed: %v", err)
	}
	sess.pickerMsgID = 0
	Send(api, chatID, nextQuestion(sess.step))
}

func (h *Handlers) applyAirport(sess *subscribeSession, code, city, label string) (confirm string, ok bool) {
	switch sess.step {
	case stepFrom:
		sess.fromIATA = code
		sess.fromCity = city
		sess.step = stepTo
		return "Откуда: " + label, true
	case stepTo:
		if code == sess.fromIATA {
			return "", false
		}
		sess.toIATA = code
		sess.toCity = city
		sess.step = stepDateFrom
		return "Куда: " + label, true
	}
	return "", false
}

// pickerLabels switches to airport-name labels when one city holds several
// airports in the list — "Москва (SVO)" three times is useless.
func pickerLabels(airports []airport.Airport) []string {
	type cityKey struct{ city, country string }
	seen := make(map[cityKey]int, len(airports))
	for _, a := range airports {
		seen[cityKey{a.City, a.Country}]++
	}
	labels := make([]string, len(airports))
	for i, a := range airports {
		if a.City != "" && seen[cityKey{a.City, a.Country}] > 1 {
			labels[i] = a.ShortLabel()
		} else {
			labels[i] = a.Label()
		}
	}
	return labels
}

func nextQuestion(s step) string {
	switch s {
	case stepTo:
		return "Куда летим? Напиши город или аэропорт."
	case stepDateFrom:
		return "С какой даты искать? Формат ДД.ММ.ГГГГ (например 01.08.2026)."
	}
	return ""
}

func (h *Handlers) removePicker(api *tgbotapi.BotAPI, chatID int64, sess *subscribeSession) {
	if sess.pickerMsgID == 0 {
		return
	}
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, sess.pickerMsgID, tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	})
	if _, err := api.Send(edit); err != nil {
		log.Printf("remove airport picker keyboard failed: %v", err)
	}
	sess.pickerMsgID = 0
}
