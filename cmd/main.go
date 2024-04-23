package main

import (
	"banner-manager/db"
	"banner-manager/internal/handlers"
	_ "fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	db.InitRedisUser()
	defer db.CloseRedisUser()

	stopCleanup := make(chan struct{})
	defer close(stopCleanup)

	go db.PeriodicallyCleanExpiredRedisTokens(time.Second, stopCleanup)

	err := db.ConnectPostgresDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.ClosePostgresDB()
	db.WaitWhileDBNotReady()

	router := mux.NewRouter()

	router.HandleFunc("/register", handlers.Register).Methods("POST")
	router.HandleFunc("/authorize", handlers.Authorize).Methods("POST")
	router.HandleFunc("/user_banner", handlers.GetUserBannerHandler).Methods("GET")
	router.HandleFunc("/banner", handlers.GetBannersByFeatureOrTagHandler).Methods("GET")
	router.HandleFunc("/banner", handlers.CreateBannerHandler).Methods("POST")
	router.HandleFunc("/banner/{id}", handlers.UpdateBannerHandler).Methods("PATCH")
	router.HandleFunc("/banner/{id}", handlers.DeleteBannerHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}
