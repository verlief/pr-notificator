package entities

type Review struct {
	Reviewer     Username    `json:"reviewer"`
	PullRequest  PullRequest `json:"pull_request"`
	ApproveCount int         `json:"approve_count"`
}
