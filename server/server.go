package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pull-request-notificator/notifier"
	"pull-request-notificator/server/entities"
)

func Run(notifier *notifier.Notifier) error {
	http.HandleFunc("/opened", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.PullRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о новом pull request: некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf("*🚀 Новый PR от* %s\n\n%s", payload.Author.Link(), payload.TextWithLink())

			if err := notifier.Send(context.Background(), message); err != nil {
				log.Printf("Ошибка отправки уведомления о новом pull request: %s", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /request-review", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.RequestReviewer
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о запросе на review: некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf("*👀 @%s, тебя приглашают на ревью*\n\n%s (by %s)", payload.Reviewer.Tag(), payload.PullRequest.TextWithLink(), payload.PullRequest.Author.Link())

			if err := notifier.Send(context.Background(), message); err != nil {
				log.Printf("Ошибка отправки уведомления о запросе на review: %s", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /approve", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о результате ревью (approve): некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf("✅ *@%s, твои изменения одобрил(а)* %s\n\n%s", payload.PullRequest.Author.Tag(), payload.Reviewer.Link(), payload.PullRequest.TextWithLink())

			if err := notifier.Send(context.Background(), message); err != nil {
				log.Printf("Ошибка отправки уведомления о результате ревью (approve): %v", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /request-changes", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о результате ревью (request-changes): некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf("❌ *@%s, тебя просит внести изменения* %s\n\n%s", payload.PullRequest.Author.Tag(), payload.Reviewer.Link(), payload.PullRequest.TextWithLink())

			if err := notifier.Send(context.Background(), message); err != nil {
				http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о результате ревью (request-changes): %v", err), http.StatusInternalServerError)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /comment", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о результате ревью (comment): некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		if payload.PullRequest.Author != payload.Reviewer {
			go func() {
				message := fmt.Sprintf("*✍️ @%s, тебе оставил(а) комментарий* %s\n\n%s", payload.PullRequest.Author.Tag(), payload.Reviewer.Link(), payload.PullRequest.TextWithLink())

				if err := notifier.Send(context.Background(), message); err != nil {
					http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о результате ревью (comment): %v", err), http.StatusInternalServerError)
				}
			}()
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /rspec-fail", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.PullRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о проваленных тестах: некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf("*🤒 @%s, возникли ошибки во время прогона тестов на CI*\n\n%s", payload.Author.Tag(), payload.TextWithLink())

			if err := notifier.Send(context.Background(), message); err != nil {
				http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о проваленных тестах: %v", err), http.StatusInternalServerError)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /rubocop-fail", func(w http.ResponseWriter, r *http.Request) {
		var payload entities.PullRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Ошибка отправки уведомления о проваленных тестах: некорректные параметры запроса: %s\n", err)
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf("*🤖 @%s, rubocop обнаружил проблемы в твоем коде*\n\n%s", payload.Author.Tag(), payload.TextWithLink())

			if err := notifier.Send(context.Background(), message); err != nil {
				http.Error(w, fmt.Sprintf("Ошибка отправки уведомления об ошбках линтера: %v", err), http.StatusInternalServerError)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Сервер запущен на :8080")
	return http.ListenAndServe("0.0.0.0:8080", nil)
}
