package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const perUpdateTimeout = 15 * time.Second

func Run(ctx context.Context, token string, subs *subscription.Repository) error {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("bot init: %w", err)
	}
	log.Printf("bot authorized as @%s", api.Self.UserName)

	router := NewRouter()
	RegisterHandlers(router, NewHandlers(subs))

	if err := registerCommandMenu(api); err != nil {
		log.Printf("set command menu failed: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := api.GetUpdatesChan(u)

	go func() {
		<-ctx.Done()
		api.StopReceivingUpdates()
	}()

	for update := range updates {
		if update.Message == nil {
			continue
		}
		msgCtx, cancel := context.WithTimeout(ctx, perUpdateTimeout)
		router.Dispatch(msgCtx, api, update.Message)
		cancel()
	}

	return ctx.Err()
}

func registerCommandMenu(api *tgbotapi.BotAPI) error {
	cmds := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "subscribe", Description: "Подписаться на цену рейса"},
		tgbotapi.BotCommand{Command: "list", Description: "Мои подписки"},
		tgbotapi.BotCommand{Command: "unsubscribe", Description: "Удалить подписку"},
		tgbotapi.BotCommand{Command: "help", Description: "Помощь"},
	)
	_, err := api.Request(cmds)
	return err
}
