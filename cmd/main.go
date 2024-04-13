package main

import (
    "banner-manager/internal/handlers"
	"banner-manager/db"
	"log"
	"net/http"
    "github.com/gorilla/mux"
)

func main() {
    db.InitRedis()
    defer db.CloseRedis()

	err := db.ConnectPostgresDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.ClosePostgresDB()

	router := mux.NewRouter()

    router.HandleFunc("/register", handlers.Register).Methods("POST")
    router.HandleFunc("/authorize", handlers.Authorize).Methods("POST")
	router.HandleFunc("/user_banner", handlers.GetUserBannerHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
