package main

import (
	"backend-task/internal/handlers"
	pullreq "backend-task/internal/pullrequest"
	"backend-task/internal/service"
	"backend-task/internal/user"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error with loading config: %s", err)
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("No database connection: %s", err)
	}

	if err = conn.Ping(context.Background()); err != nil {
		log.Fatalf("Database is not responding: %s", err)
	}

	usersRepo := user.NewRepo(conn, context.Background())
	pullreqRepo := pullreq.NewRepo(conn, context.Background())

	serv := service.NewService(pullreqRepo, usersRepo)

	handlers := handlers.Handlers{
		Serv: serv,
	}

	router := mux.NewRouter()
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
