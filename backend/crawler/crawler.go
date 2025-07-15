package crawler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/payal8797/sykell-task/backend/db"
)

// CrawlURL is the main function that:
// - Downloads the page HTML
// - Extracts metadata (title, headings, links, etc.)
// - Checks for broken links
// - Detects login forms
// - Updates all info into the database
func CrawlURL(urlStr string, id int64) {
	log.Println("🌐 Crawling:", urlStr)

	// Fetch HTML from the provided URL
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Println("❌ HTTP error:", err)
		updateStatus(id, "error")
		return
	}
	defer resp.Body.Close()

	// Load HTML into goquery for parsing
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("❌ goquery error:", err)
		updateStatus(id, "error")
		return
	}

	// -------------------------------------
	// 1. Extract <title> of the page
	title := strings.TrimSpace(doc.Find("title").First().Text())

	// -------------------------------------
	// 2. Count heading tags (H1–H6)
	headings := map[string]int{}
	for i := 1; i <= 6; i++ {
		tag := "h" + string('0'+i) // e.g., h1, h2...
		headings[tag] = doc.Find(tag).Length()
	}
	headingsJSON, _ := json.Marshal(headings) // convert to JSON for DB

	// -------------------------------------
	// 3. Count internal and external links and detect broken links
	base, _ := url.Parse(urlStr)
	internalLinks := 0
	externalLinks := 0
	var brokenLinks []string

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		link, err := url.Parse(href)
		if err != nil || link.Scheme == "javascript" {
			return
		}

		// Resolve relative URL to absolute
		fullURL := link
		if !link.IsAbs() {
			fullURL = base.ResolveReference(link)
		}

		// Classify as internal or external
		if fullURL.Host == base.Host {
			internalLinks++
		} else {
			externalLinks++
		}

		// Check if link is broken (status code 4xx or 5xx)
		client := http.Client{Timeout: 5 * time.Second}
		res, err := client.Head(fullURL.String())
		if err != nil || res.StatusCode >= 400 {
			brokenLinks = append(brokenLinks, fullURL.String())
		}
	})
	brokenLinksJSON, _ := json.Marshal(brokenLinks)

	// -------------------------------------
	// 4. Check if page contains a login form
	loginForm := doc.Find("input[type='password']").Length() > 0

	// -------------------------------------
	// 5. Store crawl results in database
	_, err = db.DB.Exec(`
		UPDATE urls 
		SET 
			status = ?,
			html_version = ?, 
			page_title = ?, 
			headings = ?, 
			internal_links = ?, 
			external_links = ?, 
			broken_links = ?, 
			login_form_detected = ?
		WHERE id = ?`,
		"done", "HTML5", title, string(headingsJSON), internalLinks,
		externalLinks, string(brokenLinksJSON), loginForm, id,
	)
	if err != nil {
		log.Println("❌ DB update error:", err)
		return
	}

	log.Println("✅ Crawl finished for:", urlStr)
}

// updateStatus is a helper to update the status column for a URL record (e.g., 'error', 'queued')
func updateStatus(id int64, status string) {
	_, err := db.DB.Exec(`UPDATE urls SET status=? WHERE id=?`, status, id)
	if err != nil {
		log.Println("❌ Failed to update status:", err)
	}
}
