package models

type Banner struct {
	ID      int         `json:"id"`
	Content interface{} `json:"content"`
	Feature int         `json:"feature"`
	Tags    []int       `json:"tags"`
}

func NewBanner(id int, content interface{}, feature int, tags []int) *Banner {
	return &Banner{
		ID:      id,
		Content: content,
		Feature: feature,
		Tags:    tags,
	}
}