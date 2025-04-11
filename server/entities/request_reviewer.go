package entities

type RequestReviewer struct {
	Requester   Username    `json:"reqester"`
	Reviewer    Username    `json:"reviewer"`
	PullRequest PullRequest `json:"pull_request"`
}
