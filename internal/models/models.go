package models

import (
	"errors"
	"time"
)

var (
	ERROR_TEAM_EXISTS  = errors.New("team_name already exists")
	ERROR_NOT_FOUND    = errors.New("resource not found")
	ERROR_PR_EXISTS    = errors.New("PR id already exists")
	ERROR_PR_MERGED    = errors.New("cannot reassign on merged PR")
	ERROR_NOT_ASSIGNED = errors.New("reviewer is not assigned to this PR")
	ERROR_NO_CANDIDATE = errors.New("no active replacement candidate in team")
)

type User struct {
	Id       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name,omitempty"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string `json:"team_name"`
	Members  []User `json:"members"`
}

type PullRequest struct {
	Id                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorId          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
	OldReviewerId     string     `json:"old_reviewer_id,omitempty"`
}

type NewPullRequest struct {
	PullRequestId  string `json:"pull_request_id"`
	PullRequstName string `json:"pull_request_name"`
	AuthorId       string `json:"author_id"`
	OldReviewerId  string `json:"old_reviewer_id"`
}

type ErrorResponseBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorResponseBody `json:"error"`
}

type UserWithStatus struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type PRresponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by,omitempty"`
}

type UserReponse struct {
	UserId       string        `json:"user_id"`
	PullRequests []PullRequest `json:"pull_requests"`
}
