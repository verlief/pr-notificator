package entities

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type PullRequest struct {
	Title  string   `json:"title"`
	URL    string   `json:"html_url"`
	Author Username `json:"author"`
}

func (p *PullRequest) TextWithLink() string {
	var cleanTitle string
	title := p.EncodedTitle()

	cleanTitle = regexp.MustCompile(`(?i)\[draft\]`).ReplaceAllString(title, "DRAFT:")
	cleanTitle = regexp.MustCompile(`(?i)\[epic\]`).ReplaceAllString(title, "EPIC:")

	cleanTitle = strings.ReplaceAll(cleanTitle, "[", "")
	cleanTitle = strings.ReplaceAll(cleanTitle, "]", ":")

	return fmt.Sprintf("[%s](%s)", cleanTitle, p.URL)
}

func (p *PullRequest) EncodedTitle() string {
	decoded, err := base64.StdEncoding.DecodeString(p.Title)
	if err != nil {
		log.Printf("Ошибка разбора title (%s): %s", p.Title, err)

		return ""
	}

	return string(decoded)
}
