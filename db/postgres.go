package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
    "banner-manager/internal/models"
)

var db *sql.DB

func ConnectPostgresDB() error {
    connStr := "user=postgres dbname=banner_manager_tables password=1234 sslmode=disable"
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

func ClosePostgresDB() {
    if db != nil {
        db.Close()
        log.Println("Disconnected from PostgreSQL database")
    }
}

func GetPostgresDB() (*sql.DB, error) {
	err := db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetLastBannerRevisionFromDB(tagID, featureID string) (*models.Banner, error) {
    var title, text, url string
	err := db.QueryRow("SELECT title, text, url FROM banners WHERE tag_id = $1 AND feature_id = $2 ORDER BY updated_at DESC LIMIT 1", tagID, featureID).Scan(&title, &text, &url)
	if err != nil {
		return nil, err
	}

	banner := &models.Banner{
		Title: title,
		Text:  text,
		URL:   url,
	}

	return banner, nil
}