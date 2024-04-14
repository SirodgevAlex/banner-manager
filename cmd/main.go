package main

import (
	"banner-manager/db"
	"banner-manager/internal/handlers"
	"fmt"
	"log"
	"net/http"
	_ "time"

	"github.com/gorilla/mux"
)

func main() {
	db.InitRedisUser()
	defer db.CloseRedisUser()

	// stopCleanup := make(chan struct{})
	// defer close(stopCleanup)

	// go db.PeriodicallyCleanExpiredRedisTokens(time.Second, stopCleanup) пока забъем

	err := db.ConnectPostgresDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.ClosePostgresDB()

	router := mux.NewRouter()

	fmt.Println(handlers.IsAdminTokenValid("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozLCJpc19hZG1pbiI6dHJ1ZSwiZXhwIjoxNzEzMDM3Mjc5LCJzdWIiOiIzIn0.z-EruTrVpKvPI0qcshydkbhEDKRyNk-UswkCx2pT8MY"))

	router.HandleFunc("/register", handlers.Register).Methods("POST")
	router.HandleFunc("/authorize", handlers.Authorize).Methods("POST")
	router.HandleFunc("/user_banner", handlers.GetUserBannerHandler).Methods("GET") //протестировать 2 и понять, что там с uselastrevision
	router.HandleFunc("/banner", handlers.GetBannersByFeatureOrTagHandler).Methods("GET")
	router.HandleFunc("/banner", handlers.CreateBannerHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
