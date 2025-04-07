package main

import (
	"log"
	"os"
	"pull-request-notificator/notifier"
	"pull-request-notificator/server"
	"strconv"
)

func main() {
	var threadID int64 = 0

	tokenString := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatIDString := os.Getenv("TELEGRAM_CHAT_ID")
	threadIDString := os.Getenv("TELEGRAM_THREAD_ID")

	if tokenString == "" || chatIDString == "" {
		log.Printf("Отсутствуют переменные окружения TELEGRAM_BOT_TOKEN или TELEGRAM_CHAT_ID")
		return
	}

	chatID, err := strconv.ParseInt(chatIDString, 10, 64)
	if err != nil {
		log.Printf("Ошибка преобразования TELEGRAM_CHAT_ID: %v", err)
		return
	}

	if threadIDString != "" {
		if threadID, err = strconv.ParseInt(threadIDString, 10, 64); err != nil {
			log.Printf("Ошибка преобразования TELEGRAM_THREAD_ID: %v", err)
			return
		}
	}

	notifier, err := notifier.New(tokenString, chatID, threadID)
	if err != nil {
		log.Fatalf("Ошибка инициализации сервиса: %v", err)
	}

	server.Run(notifier)
}
