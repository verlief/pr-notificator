package entities

import (
	"log"
	"strconv"
)

type Review struct {
	Reviewer     Username    `json:"reviewer"`
	PullRequest  PullRequest `json:"pull_request"`
	ApproveCount string      `json:"approve_count"`
}

func (r *Review) ApproveCountAsInt() int {
	count, err := strconv.Atoi(r.ApproveCount)
	if err != nil {
		log.Printf("Ошибка преобразования approve_count (%s): %s", r.ApproveCount, err)

		return 1
	}
	return count
}
