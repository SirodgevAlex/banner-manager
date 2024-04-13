package models

import "time"

type Banner struct {
	ID        int       `json:"id"`
	FeatureID int       `json:"feature_id"`
	TagID     int       `json:"tag_id"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	URL       string    `json:"url"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewBanner(id int, title, text, url string, featureID, tagID int, isActive bool, createdAt, updatedAt time.Time) *Banner {
	return &Banner{
		ID:        id,
		Title:     title,
		Text:      text,
		URL:       url,
		FeatureID: featureID,
		TagID:     tagID,
		IsActive:  isActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
