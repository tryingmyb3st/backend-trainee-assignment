package main

import (
	"backend-task/internal/handlers"
	"backend-task/internal/pullrequest"
	"backend-task/internal/teams"
	"backend-task/internal/user"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	teamsRepo := teams.NewTeamsRepo()
	usersRepo := user.NewUsersRepo()
	prRepo := pullrequest.NewPrRepo()
	handlers := handlers.Handlers{
		TeamsRepo: teamsRepo,
		UsersRepo: usersRepo,
		PRrepo:    prRepo,
	}

	router.HandleFunc("/team/add", handlers.HandleAddNewTeam).Methods("POST")
	router.HandleFunc("/team/get", handlers.HandleGetTeam).Queries("team_name", "{team_name}").Methods("GET")
	router.HandleFunc("/users/setIsActive", handlers.HandleSetIsActiveUser).Methods("POST")
	router.HandleFunc("/users/getReview", handlers.HandleGetUserReview).Queries("user_id", "{user_id}").Methods("GET")
	router.HandleFunc("/pullRequest/create", handlers.HandleCreatePullRequest).Methods("POST")
	router.HandleFunc("/pullRequest/merge", handlers.HandleMergePullRequest).Methods("POST")
	router.HandleFunc("/pullRequest/reassign", handlers.HandleReassignUser).Methods("POST")

	log.Println("Starting server at :8080")
	http.ListenAndServe(":8080", router)
}
