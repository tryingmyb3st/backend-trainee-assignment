package handlers

import (
	"backend-task/internal/pullrequest"
	"backend-task/internal/teams"
	"backend-task/internal/user"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Handlers struct {
	PRrepo    *pullrequest.PRrepository
	UsersRepo *user.UsersRepository
	TeamsRepo *teams.TeamsRepository
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
	PR         pullrequest.PullRequest `json:"pr"`
	MergedAt   *time.Time              `json:"mergedAt,omitempty"`
	ReplacedBy string                  `json:"replaced_by,omitempty"`
}

type UserReponse struct {
	UserId       string                    `json:"user_id"`
	PullRequests []pullrequest.PullRequest `json:"pull_requests"`
}

func (h *Handlers) HandleAddNewTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	team := teams.Team{}
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	createdTeam, err := h.TeamsRepo.CreateNewTeam(team)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "TEAM_EXISTS",
				Message: err.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	h.UsersRepo.AddNewUsers(createdTeam.TeamName, createdTeam.Members)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdTeam); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleGetTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	teamName := r.URL.Query().Get("team_name")

	team, err := h.TeamsRepo.GetTeamByName(teamName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	if err := json.NewEncoder(w).Encode(team); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleSetIsActiveUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userWithStatus := UserWithStatus{}
	if err := json.NewDecoder(r.Body).Decode(&userWithStatus); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	user, err := h.UsersRepo.ChangeUserStatus(userWithStatus.UserId, userWithStatus.IsActive)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}
	h.TeamsRepo.ChangeTeamMemberStatus(user.TeamName, user.Id, user.IsActive)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleGetUserReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userId := r.URL.Query().Get("user_id")

	PRs, err := h.PRrepo.GetUserPRs(userId, h.UsersRepo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: teams.ERROR_NOT_FOUND.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	resp := UserReponse{
		UserId:       userId,
		PullRequests: PRs,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleCreatePullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	newPR := pullrequest.NewPullRequest{}
	if err := json.NewDecoder(r.Body).Decode(&newPR); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	exist := h.UsersRepo.IsUserExists(newPR.AuthorId)
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: teams.ERROR_NOT_FOUND.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	pr, err := h.PRrepo.CreateNewPR(newPR, h.UsersRepo, h.TeamsRepo)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "PR_EXISTS",
				Message: err.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	resp := PRresponse{
		PR: pr,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleMergePullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pr := pullrequest.NewPullRequest{}
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	mergedPR, err := h.PRrepo.MergePR(pr, h.UsersRepo, h.TeamsRepo)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	resp := PRresponse{
		PR:       mergedPR,
		MergedAt: &mergedPR.MergedAt,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleReassignUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pr := pullrequest.NewPullRequest{}
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	exist := h.UsersRepo.IsUserExists(pr.OldReviewerId)
	if !exist {
		w.WriteHeader(http.StatusNotFound)
		resp := ErrorResponse{
			Error: ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: teams.ERROR_NOT_FOUND.Error(),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	modifiedPR, repalacedBy, err := h.PRrepo.ReassignUser(pr, h.UsersRepo, h.TeamsRepo)
	if err != nil {
		if errors.Is(err, pullrequest.ERROR_PR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp := ErrorResponse{
				Error: ErrorResponseBody{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
		} else {
			w.WriteHeader(http.StatusConflict)
			resp := ErrorResponse{
				Error: ErrorResponseBody{
					Code:    "PR_MERGED",
					Message: err.Error(),
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}

		}
		return
	}
	resp := PRresponse{
		PR:         modifiedPR,
		ReplacedBy: repalacedBy,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}
