package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message)

type CallbackFunc func(ctx context.Context, api *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery)

type Interceptor func(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) bool

type Router struct {
	commands    map[string]HandlerFunc
	callbacks   map[string]CallbackFunc
	fallback    HandlerFunc
	interceptor Interceptor
}

func NewRouter() *Router {
	return &Router{
		commands:  make(map[string]HandlerFunc),
		callbacks: make(map[string]CallbackFunc),
	}
}

func (r *Router) Handle(command string, h HandlerFunc) {
	r.commands[command] = h
}

// HandleCallback registers a handler for inline-button callbacks whose data
// starts with "<prefix>:".
func (r *Router) HandleCallback(prefix string, h CallbackFunc) {
	r.callbacks[prefix] = h
}

func (r *Router) Fallback(h HandlerFunc) {
	r.fallback = h
}

func (r *Router) Intercept(i Interceptor) {
	r.interceptor = i
}

func (r *Router) Dispatch(ctx context.Context, api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if r.interceptor != nil && r.interceptor(ctx, api, msg) {
		return
	}
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

func (r *Router) DispatchCallback(ctx context.Context, api *tgbotapi.BotAPI, cq *tgbotapi.CallbackQuery) {
	prefix, _, _ := strings.Cut(cq.Data, ":")
	if h, ok := r.callbacks[prefix]; ok {
		h(ctx, api, cq)
	}
}
