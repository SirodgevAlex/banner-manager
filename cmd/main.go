package main

import (
	"banner-manager/db"
	"banner-manager/internal/handlers"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
    db.InitRedis()
    defer db.CloseRedis()
    db.PeriodicallyCleanExpiredRedisTokens(time.Second)

	err := db.ConnectPostgresDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.ClosePostgresDB()

	router := mux.NewRouter()

    router.HandleFunc("/register", handlers.Register).Methods("POST")
    router.HandleFunc("/authorize", handlers.Authorize).Methods("POST")
	router.HandleFunc("/user_banner", handlers.GetUserBannerHandler).Methods("GET")
    router.HandleFunc("/banner", handlers.CreateBannerHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8085", router))
}
