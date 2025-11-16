package pullrequest

import "time"

const (
	OPEN   = "OPEN"
	MERGED = "MERGED"
)

type PullRequest struct {
	Id                string    `json:"pull_request_id"`
	Name              string    `json:"pull_request_name"`
	AuthorId          string    `json:"author_id"`
	Status            string    `json:"status"`
	AssignedReviewers []string  `json:"assigned_reviewers,omitempty"`
	MergedAt          time.Time `json:"-"`
}

type NewPullRequest struct {
	PullRequestId  string `json:"pull_request_id"`
	PullRequstName string `json:"pull_request_name"`
	AuthorId       string `json:"author_id"`
	OldReviewerId  string `json:"old_reviewer_id"`
}
