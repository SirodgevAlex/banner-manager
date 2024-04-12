package main

import (
    "banner-manager/internal/handlers"
	"banner-manager/db"
	"log"
	"net/http"
    "github.com/gorilla/mux"
)

func main() {
	err := db.ConnectPostgresDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.ClosePostgresDB()

	router := mux.NewRouter()

    router.HandleFunc("/register", handlers.Register).Methods("POST")
	//router.HandleFunc("/user_banner", ).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
