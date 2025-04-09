package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pull-request-notificator/notifier"
	"regexp"

	"gopkg.in/yaml.v3"
)

type PullRequest struct {
	Title  string `json:"title"`
	URL    string `json:"html_url"`
	Author string `json:"author"`
}

type RequestReviewer struct {
	Requester   string      `json:"reqester"`
	Reviewer    string      `json:"reviewer"`
	PullRequest PullRequest `json:"pull_request"`
}

type Review struct {
	Reviewer    string      `json:"reviewer"`
	PullRequest PullRequest `json:"pull_request"`
}

var username_mapper map[string]string = nil

func Run(notifier *notifier.Notifier) error {
	http.HandleFunc("/opened", func(w http.ResponseWriter, r *http.Request) {
		var payload PullRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println("Ошибка отправки уведомления о новом pull request: некорректные параметры запроса")
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		// Запускаем горутину для уведомления
		go func() {
			message := fmt.Sprintf(
				"*🚀 Новый PR от* %s\n\n%s",
				usernameAsLink(resolveUsername(payload.Author)),
				pullRequestLink(payload.Title, payload.URL),
			)

			if err := notifier.Send(context.Background(), message); err != nil {
				log.Printf("Ошибка отправки уведомления о новом pull request: %s", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /request-review", func(w http.ResponseWriter, r *http.Request) {
		var payload RequestReviewer
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println("Ошибка отправки уведомления о запросе на review: некорректные параметры запроса")
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf(
				"*👀 @%s, тебя приглашают на ревью*\n\n%s (by %s)",
				resolveUsername(payload.Reviewer),
				pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
				usernameAsLink(resolveUsername(payload.PullRequest.Author)),
			)

			if err := notifier.Send(context.Background(), message); err != nil {
				log.Printf("Ошибка отправки уведомления о запросе на review: %s", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /approve", func(w http.ResponseWriter, r *http.Request) {
		var payload Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println("Ошибка отправки уведомления о результате ревью (approve): некорректные параметры запроса")
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf(
				"✅ *@%s, твои изменения одобрил(а)* %s\n\n%s",
				resolveUsername(payload.PullRequest.Author),
				usernameAsLink(resolveUsername(payload.Reviewer)),
				pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
			)

			if err := notifier.Send(context.Background(), message); err != nil {
				log.Printf("Ошибка отправки уведомления о результате ревью (approve): %v", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /request-changes", func(w http.ResponseWriter, r *http.Request) {
		var payload Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println("Ошибка отправки уведомления о результате ревью (request-changes): некорректные параметры запроса")
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		go func() {
			message := fmt.Sprintf(
				"❌ *@%s, тебя просит внести изменения* %s\n\n%s",
				resolveUsername(payload.PullRequest.Author),
				usernameAsLink(resolveUsername(payload.Reviewer)),
				pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
			)

			if err := notifier.Send(context.Background(), message); err != nil {
				http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о результате ревью (request-changes): %v", err), http.StatusInternalServerError)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("POST /comment", func(w http.ResponseWriter, r *http.Request) {
		var payload Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println("Ошибка отправки уведомления о результате ревью (comment): некорректные параметры запроса")
			http.Error(w, "Invalid request parameters", http.StatusBadRequest)
			return
		}

		if payload.PullRequest.Author != payload.Reviewer {
			go func() {
				message := fmt.Sprintf(
					"*✍️ @%s, тебе оставил(а) комментарий* %s\n\n%s",
					resolveUsername(payload.PullRequest.Author),
					usernameAsLink(resolveUsername(payload.Reviewer)),
					pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
				)

				if err := notifier.Send(context.Background(), message); err != nil {
					http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о результате ревью (comment): %v", err), http.StatusInternalServerError)
				}
			}()
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Сервер запущен на :8080")
	return http.ListenAndServe("0.0.0.0:8080", nil)
}

func usernameAsLink(username string) string {
	return fmt.Sprintf("[@%s](tg://resolve?domain=%s)", username, username)
}

func pullRequestLink(title, url string) string {
	re := regexp.MustCompile(`(?i)\[draft\]`)
	cleanTitle := re.ReplaceAllString(title, "DRAFT:")
	return fmt.Sprintf("[%s](%s)", cleanTitle, url)
}

func resolveUsername(target_username string) string {
	var err error
	if username_mapper == nil {
		username_mapper, err = parseYAML()
		if err != nil {
			log.Printf("Не удалось спарсить yaml: %s", err)

			return target_username
		}
	}

	username, ok := username_mapper[target_username]
	if !ok {
		return target_username
	}

	return username
}

func parseYAML() (map[string]string, error) {
	filename := os.Getenv("GITHUB_USERNAME_MAPPER")
	if filename == "" {
		return nil, fmt.Errorf("Отсутствуют переменная окружения GITHUB_USERNAME_MAPPER")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := yaml.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}

	return result, nil
}
