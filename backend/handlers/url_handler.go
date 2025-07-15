package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/payal8797/sykell-task/backend/crawler"
	"github.com/payal8797/sykell-task/backend/db"
	"github.com/payal8797/sykell-task/backend/models"
)

// Base SQL query reused in multiple functions
const baseQuery = `
SELECT id, url, status, html_version, page_title, headings,
       internal_links, external_links, broken_links,
       login_form_detected, created_at
FROM urls`

// mapToResponse converts DB model (with sql.Null types) to clean JSON-friendly struct
func mapToResponse(u models.URL) models.URLResponse {
	return models.URLResponse{
		ID:                u.ID,
		URL:               u.URL,
		Status:            u.Status,
		HTMLVersion:       nullToPtr(u.HTMLVersion),
		PageTitle:         nullToPtr(u.PageTitle),
		Headings:          nullToPtr(u.Headings),
		InternalLinks:     u.InternalLinks,
		ExternalLinks:     u.ExternalLinks,
		BrokenLinks:       nullToPtr(u.BrokenLinks),
		LoginFormDetected: u.LoginFormDetected,
		CreatedAt:         u.CreatedAt,
	}
}

// PostURL handles POST /urls
// Accepts a URL input, saves it in DB, triggers crawler, and returns the record
func PostURL(c *gin.Context) {
	var input struct {
		URL string `json:"url" binding:"required"`
	}

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL is required"})
		return
	}

	// Insert new URL with status = 'queued'
	result, err := db.DB.Exec(`
		INSERT INTO urls (url, status, internal_links, external_links)
		VALUES (?, 'queued', 0, 0)`, input.URL)
	if err != nil {
		log.Println("❌ DB insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get inserted row's ID
	id, _ := result.LastInsertId()

	// Trigger background crawling
	go crawler.CrawlURL(input.URL, id)

	// Fetch full record to return in response
	url, err := fetchURLByID(id)
	if err != nil {
		log.Println("❌ Fetch after insert failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fetch failed"})
		return
	}

	c.JSON(http.StatusOK, mapToResponse(url))
}

// GetAllURLs handles GET /urls
// Fetches and returns all crawled URL records from the database
func GetAllURLs(c *gin.Context) {
	rows, err := db.DB.Query(baseQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer rows.Close()

	var urls []models.URLResponse
	for rows.Next() {
		var u models.URL
		// Scan each row
		if err := rows.Scan(&u.ID, &u.URL, &u.Status, &u.HTMLVersion, &u.PageTitle,
			&u.Headings, &u.InternalLinks, &u.ExternalLinks,
			&u.BrokenLinks, &u.LoginFormDetected, &u.CreatedAt); err != nil {
			log.Println("Row scan error:", err)
			continue
		}
		urls = append(urls, mapToResponse(u))
	}

	c.JSON(http.StatusOK, urls)
}

// GetURLByID handles GET /urls/:id
// Returns the detailed result for a specific URL by ID
func GetURLByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	url, err := fetchURLByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.JSON(http.StatusOK, mapToResponse(url))
}

// ReanalyzeURL handles POST /urls/:id/reanalyze
// Updates status to 'queued' and re-triggers the crawler for the same URL
func ReanalyzeURL(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Fetch existing URL from DB
	var url string
	err = db.DB.QueryRow("SELECT url FROM urls WHERE id = ?", id).Scan(&url)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Reset status to 'queued'
	_, err = db.DB.Exec("UPDATE urls SET status = 'queued' WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// Launch crawling again in background
	go crawler.CrawlURL(url, id)

	c.JSON(http.StatusOK, gin.H{"message": "Reanalysis started"})
}

// DeleteURL handles DELETE /urls/:id
// Deletes the record for a given URL
func DeleteURL(c *gin.Context) {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM urls WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// fetchURLByID is a helper to get full row from DB by ID
func fetchURLByID(id int64) (models.URL, error) {
	var u models.URL
	err := db.DB.QueryRow(baseQuery+" WHERE id = ?", id).
		Scan(&u.ID, &u.URL, &u.Status, &u.HTMLVersion, &u.PageTitle,
			&u.Headings, &u.InternalLinks, &u.ExternalLinks, &u.BrokenLinks,
			&u.LoginFormDetected, &u.CreatedAt)
	return u, err
}

// nullToPtr converts sql.NullString to *string for clean JSON output
func nullToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
