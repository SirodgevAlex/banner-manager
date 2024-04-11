package postgres

import (
    "database/sql"
    "log"
    _ "github.com/lib/pq"
)

var db *sql.DB

func ConnectDB() error {
    connStr := "user=postgres dbname=banner-manager-db password=1234 sslmode=disable"
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    err = db.Ping()
    if err != nil {
        return err
    }
    log.Println("Connected to PostgreSQL database")
    return nil
}

func CloseDB() {
    if db != nil {
        db.Close()
        log.Println("Disconnected from PostgreSQL database")
    }
}
