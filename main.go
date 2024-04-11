package main

import (
    "log"
    "net/http"
    "banner-manager/api"
    "banner-manager/db/postgres"
)

func main() {
    err := postgres.ConnectDB()
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }
    defer postgres.CloseDB()

    router := api.NewRouter()

    addr := ":8080"
    if err := http.ListenAndServe(addr, router); err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
}
