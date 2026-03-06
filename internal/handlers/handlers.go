package handlers

import (
	"backend-task/internal/models"
	"backend-task/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Handlers struct {
	Serv service.Service
}

func (h *Handlers) HandleAddNewTeam(w http.ResponseWriter, r *http.Request) {
	log.Println("handling HandleAddNewTeam...")
	w.Header().Set("Content-Type", "application/json")

	team := models.Team{}

	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	err := h.Serv.CreateNewTeam(team)
	if err != nil {
		if errors.Is(err, models.ERROR_TEAM_EXISTS) {
			w.WriteHeader(http.StatusBadRequest)
			resp := models.ErrorResponse{
				Error: models.ErrorResponseBody{
					Code:    "TEAM_EXISTS",
					Message: err.Error(),
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp := models.ErrorResponse{
				Error: models.ErrorResponseBody{
					Code:    "SERVER_ERROR",
					Message: "Something went wrong...",
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
			return
		}

		return
	}

	createdTeam, err := h.Serv.GetTeamByName(team.TeamName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdTeam); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleGetTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	teamName := r.URL.Query().Get("team_name")

	team, err := h.Serv.GetTeamByName(teamName)
	if err != nil {
		if errors.Is(err, models.ERROR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp := models.ErrorResponse{
				Error: models.ErrorResponseBody{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp := models.ErrorResponse{
				Error: models.ErrorResponseBody{
					Code:    "SERVER_ERROR",
					Message: "Something went wrong...",
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
		}
		return
	}

	if err := json.NewEncoder(w).Encode(team); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleSetIsActiveUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userWithStatus := models.UserWithStatus{}
	if err := json.NewDecoder(r.Body).Decode(&userWithStatus); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	err := h.Serv.ChangeUserStatus(userWithStatus)
	if err != nil {
		if errors.Is(err, models.ERROR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp := models.ErrorResponse{
				Error: models.ErrorResponseBody{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp := models.ErrorResponse{
				Error: models.ErrorResponseBody{
					Code:    "SERVER_ERROR",
					Message: "Something went wrong...",
				},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Fatal("Error with writing response")
			}
		}
		return
	}

	changedUser, err := h.Serv.GetUserByID(userWithStatus.UserId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	resp := map[string]interface{}{
		"user": changedUser,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleGetUserReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userId := r.URL.Query().Get("user_id")

	PRs, err := h.Serv.GetUserPrs(userId)
	if err != nil {
		resp := models.ErrorResponse{}

		if errors.Is(err, models.ERROR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp.Error = models.ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Error = models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			}
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	resp := models.UserReponse{
		UserId:       userId,
		PullRequests: PRs,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleCreatePullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	newPR := models.PullRequest{}
	if err := json.NewDecoder(r.Body).Decode(&newPR); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	err := h.Serv.CreateNewPR(newPR)
	if err != nil {
		resp := models.ErrorResponse{}

		if errors.Is(err, models.ERROR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp.Error = models.ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			}
		} else if errors.Is(err, models.ERROR_PR_EXISTS) {
			w.WriteHeader(http.StatusConflict)
			resp.Error = models.ErrorResponseBody{
				Code:    "PR_EXISTS",
				Message: err.Error(),
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Error = models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			}
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	addedPR, err := h.Serv.GetPRByID(newPR.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
	}

	if err := json.NewEncoder(w).Encode(addedPR); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleMergePullRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pr := models.PullRequest{}
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	err := h.Serv.MergePullRequest(pr.Id)
	if err != nil {
		resp := models.ErrorResponse{}

		if errors.Is(err, models.ERROR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp.Error = models.ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Error = models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			}
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	changedPR, err := h.Serv.GetPRByID(pr.Id)
	resp := models.PRresponse{
		PR: *changedPR,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}

func (h *Handlers) HandleReassignUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pr := models.PullRequest{}
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := models.ErrorResponse{
			Error: models.ErrorResponseBody{
				Code:    "InternalServerError",
				Message: fmt.Sprintf("Error with decoding request:  %s", err.Error()),
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	replacedBy, err := h.Serv.ReassignNewReviewer(pr)
	if err != nil {
		resp := models.ErrorResponse{}

		if errors.Is(err, models.ERROR_NOT_FOUND) {
			w.WriteHeader(http.StatusNotFound)
			resp.Error = models.ErrorResponseBody{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			}
		} else if errors.Is(err, models.ERROR_PR_MERGED) {
			w.WriteHeader(http.StatusConflict)
			resp.Error = models.ErrorResponseBody{
				Code:    "PR_MERGED",
				Message: err.Error(),
			}
		} else if errors.Is(err, models.ERROR_NOT_ASSIGNED) {
			w.WriteHeader(http.StatusConflict)
			resp.Error = models.ErrorResponseBody{
				Code:    "NOT_ASSIGNED",
				Message: err.Error(),
			}
		} else if errors.Is(err, models.ERROR_NO_CANDIDATE) {
			w.WriteHeader(http.StatusConflict)
			resp.Error = models.ErrorResponseBody{
				Code:    "NO_CANDIDATE",
				Message: err.Error(),
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Error = models.ErrorResponseBody{
				Code:    "SERVER_ERROR",
				Message: "Something went wrong...",
			}
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Fatal("Error with writing response")
		}
		return
	}

	modifiedPR, err := h.Serv.GetPRByID(pr.Id)
	resp := models.PRresponse{
		PR:         *modifiedPR,
		ReplacedBy: replacedBy.Id,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Fatal("Error with writing response")
	}
}
