package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"bilecik/internal/airport"
	"bilecik/internal/subscription"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const perUpdateTimeout = 15 * time.Second

func Run(ctx context.Context, token string, subs *subscription.Repository, airports *airport.Repository) error {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("bot init: %w", err)
	}
	log.Printf("bot authorized as @%s", api.Self.UserName)

	router := NewRouter()
	RegisterHandlers(router, NewHandlers(subs, airports))

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
		switch {
		case update.Message != nil:
			msgCtx, cancel := context.WithTimeout(ctx, perUpdateTimeout)
			router.Dispatch(msgCtx, api, update.Message)
			cancel()
		case update.CallbackQuery != nil:
			cbCtx, cancel := context.WithTimeout(ctx, perUpdateTimeout)
			router.DispatchCallback(cbCtx, api, update.CallbackQuery)
			cancel()
		}
	}

	return ctx.Err()
}

func registerCommandMenu(api *tgbotapi.BotAPI) error {
	cmds := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "subscribe", Description: "Подписаться на цену рейса"},
		tgbotapi.BotCommand{Command: "list", Description: "Мои подписки"},
		tgbotapi.BotCommand{Command: "unsubscribe", Description: "Удалить подписку"},
		tgbotapi.BotCommand{Command: "cancel", Description: "Прервать текущий диалог"},
		tgbotapi.BotCommand{Command: "help", Description: "Помощь"},
	)
	_, err := api.Request(cmds)
	return err
}
