package model

import "time"

type Tsundoku struct {
	ID           int
	UserID       int
	Category     string
	Title        string
	Author       string
	URL          string
	Deadline     time.Time
	RequiredTime string
	CreatedAt    time.Time
}

type User struct {
	DisplayName   string
	UserID        string
	Language      string
	PictureURL    string
	StatusMessage string
}
type Book struct {
	Title  string
	Author string
}
