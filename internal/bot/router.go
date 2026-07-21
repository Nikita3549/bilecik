package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message)

type Router struct {
	commands map[string]HandlerFunc
	fallback HandlerFunc
}

func NewRouter() *Router {
	return &Router{commands: make(map[string]HandlerFunc)}
}

func (r *Router) Handle(command string, h HandlerFunc) {
	r.commands[command] = h
}

func (r *Router) Fallback(h HandlerFunc) {
	r.fallback = h
}

func (r *Router) Dispatch(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.IsCommand() {
		if h, ok := r.commands[msg.Command()]; ok {
			h(ctx, api, msg)
			return
		}
	}
	if r.fallback != nil {
		r.fallback(ctx, api, msg)
	}
}
