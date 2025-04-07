package notifier

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Notifier struct {
	client   *tgbotapi.BotAPI
	chatID   int64
	threadID int64
}

func New(token string, chatID, threadID int64) (*Notifier, error) {
	client, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Notifier{
		client:   client,
		chatID:   chatID,
		threadID: threadID,
	}, nil
}

func (t Notifier) Send(ctx context.Context, message string) error {
	msg := tgbotapi.NewMessage(0, message)
	msg.ParseMode = tgbotapi.ModeMarkdown

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		msg.ChatID = t.chatID
		msg.ReplyToMessageID = int(t.threadID)

		if _, err := t.client.Send(msg); err != nil {
			return fmt.Errorf("send message to chat %d: %w", t.chatID, err)
		}
	}

	return nil
}
