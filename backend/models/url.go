package models

import (
	"database/sql"
)

// This struct is used ONLY to scan from the database
type URL struct {
	ID                int            `json:"id"`
	URL               string         `json:"url"`
	Status            string         `json:"status"`
	HTMLVersion       sql.NullString `json:"html_version"`
	PageTitle         sql.NullString `json:"page_title"`
	Headings          sql.NullString `json:"headings"`
	InternalLinks     int            `json:"internal_links"`
	ExternalLinks     int            `json:"external_links"`
	BrokenLinks       sql.NullString `json:"broken_links"`
	LoginFormDetected bool           `json:"login_form_detected"`
	CreatedAt         string         `json:"created_at"`
}

// This struct is for sending clean JSON responses to frontend
type URLResponse struct {
	ID                int     `json:"id"`
	URL               string  `json:"url"`
	Status            string  `json:"status"`
	HTMLVersion       *string `json:"html_version"`
	PageTitle         *string `json:"page_title"`
	Headings          *string `json:"headings"`
	InternalLinks     int     `json:"internal_links"`
	ExternalLinks     int     `json:"external_links"`
	BrokenLinks       *string `json:"broken_links"`
	LoginFormDetected bool    `json:"login_form_detected"`
	CreatedAt         string  `json:"created_at"`
}
