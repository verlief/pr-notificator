package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pull-request-notificator/notifier"
	"strings"

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
	http.HandleFunc("POST /opened", func(w http.ResponseWriter, r *http.Request) {
		var payload PullRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		message := fmt.Sprintf(
			"*🚀 Новый PR от* %s\n\n%s",
			usernameAsLink(resolveUsername(payload.Author)),
			pullRequestLink(payload.Title, payload.URL),
		)

		if err := notifier.Send(r.Context(), message); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о новом pull request: %v", err), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /request-review", func(w http.ResponseWriter, r *http.Request) {
		var payload RequestReviewer
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Ошибка разбора JSON", http.StatusBadRequest)
			return
		}

		message := fmt.Sprintf(
			"*👀 @%s, тебя приглашают на ревью*\n\n%s (by %s)",
			resolveUsername(payload.Reviewer),
			pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
			usernameAsLink(resolveUsername(payload.PullRequest.Author)),
		)

		if err := notifier.Send(r.Context(), message); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о запросе на review: %v", err), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /re-request-review", func(w http.ResponseWriter, r *http.Request) {
		var payload RequestReviewer
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Ошибка разбора JSON", http.StatusBadRequest)
			return
		}

		message := fmt.Sprintf(
			"*👀 @%s, тебя приглашают на ре-ревью*\n\n%s (by %s)",
			resolveUsername(payload.Reviewer),
			pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
			usernameAsLink(resolveUsername(payload.PullRequest.Author)),
		)

		if err := notifier.Send(r.Context(), message); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о запросе на review: %v", err), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /approve", func(w http.ResponseWriter, r *http.Request) {
		var payload Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Ошибка разбора JSON", http.StatusBadRequest)
			return
		}

		message := fmt.Sprintf(
			"✅ *@%s, твои изменения одобрил* %s\n\n%s",
			resolveUsername(payload.PullRequest.Author),
			usernameAsLink(resolveUsername(payload.Reviewer)),
			pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
		)

		if err := notifier.Send(r.Context(), message); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о запросе на review: %v", err), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /request-changes", func(w http.ResponseWriter, r *http.Request) {
		var payload Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Ошибка разбора JSON", http.StatusBadRequest)
			return
		}

		message := fmt.Sprintf(
			"❌ *@%s, тебя просит внести изменения* %s\n\n%s",
			resolveUsername(payload.PullRequest.Author),
			usernameAsLink(resolveUsername(payload.Reviewer)),
			pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
		)

		if err := notifier.Send(r.Context(), message); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о запросе на review: %v", err), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /comment", func(w http.ResponseWriter, r *http.Request) {
		var payload Review
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Ошибка разбора JSON", http.StatusBadRequest)
			return
		}

		if payload.PullRequest.Author != payload.Reviewer {
			message := fmt.Sprintf(
				"*✍️ @%s, тебе оставил комментарий* %s\n\n%s",
				resolveUsername(payload.PullRequest.Author),
				usernameAsLink(resolveUsername(payload.Reviewer)),
				pullRequestLink(payload.PullRequest.Title, payload.PullRequest.URL),
			)

			if err := notifier.Send(r.Context(), message); err != nil {
				http.Error(w, fmt.Sprintf("Ошибка отправки уведомления о запросе на review: %v", err), http.StatusInternalServerError)
			}
		}
	})

	log.Println("Сервер запущен на :8080")
	return http.ListenAndServe(":8080", nil)
}

func usernameAsLink(username string) string {
	return fmt.Sprintf("[@%s](tg://resolve?domain=%s)", username, username)
}

func pullRequestLink(title, url string) string {
	cleanTitle := strings.Replace(title, "[DRAFT]", "DRAFT:", 1)
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
