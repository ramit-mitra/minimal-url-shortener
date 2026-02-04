package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	profiler "github.com/blackfireio/go-continuous-profiling"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"
)

type DefaultResponse struct {
	Message string `json:"message"`
}

type Payload struct {
	URL     string `json:"url"`
	Single  *bool  `json:"single,omitempty"`
	Expires *int64 `json:"expires,omitempty"`
}

// PostgreSQL connection pool
var dbPool *pgxpool.Pool

func encodeToBase62(integer int64) string {
	const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	if integer == 0 {
		return "0"
	}

	encoded := ""
	for integer > 0 {
		remainder := integer % 62
		encoded = string(base62Chars[remainder]) + encoded
		integer /= 62
	}

	return encoded
}

// connect to PostgreSQL with connection pool
func connectToDB() {
	var err error

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ DATABASE_URL variable is not set")
	}

	dbPool, err = pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("😭 Unable to connect to database: %v\n", err)
	}

	log.Println("🎉 Connected to PostgreSQL (with connection pool)")
}

// create URLs table if it doesn't exist
func createTableIfNotExists() {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		shortcode VARCHAR(255) UNIQUE NOT NULL,
		url TEXT NOT NULL,
		single_use BOOLEAN DEFAULT FALSE,
		expires_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := dbPool.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("❌ Failed to create table: %v\n", err)
	}

	log.Println("🧑🏽‍💻 Table 'urls' is ready")
}

// save URL to PostgreSQL
func saveURLToDB(payload Payload, shortcode string) error {
	var expiresAt time.Time

	if payload.Expires != nil {
		expiresAt = time.Unix(*payload.Expires, 0)
	} else {
		// defaults to 1000 years
		yearsInSeconds := int64(1000) * 365 * 24 * 60 * 60
		expiresAt = time.Now().Add(time.Duration(yearsInSeconds) * time.Second)
	}

	singleUse := false
	if payload.Single != nil {
		singleUse = *payload.Single
	}

	_, err := dbPool.Exec(context.Background(),
		"INSERT INTO urls (shortcode, url, single_use, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		shortcode, payload.URL, singleUse, expiresAt, time.Now())
	if err != nil {
		return err
	}

	return nil
}

// Get URL from PostgreSQL
func getURLFromDB(shortcode string) (string, bool, time.Time, error) {
	var url string
	var singleUse bool
	var expiresAt time.Time

	err := dbPool.QueryRow(context.Background(),
		"SELECT url, single_use, expires_at FROM urls WHERE shortcode=$1", shortcode).Scan(&url, &singleUse, &expiresAt)
	if err != nil {
		return "", false, time.Time{}, err
	}

	return url, singleUse, expiresAt, nil
}

// delete URL from PostgreSQL
func deleteURLFromDB(shortcode string) error {
	_, err := dbPool.Exec(context.Background(), "DELETE FROM urls WHERE shortcode=$1", shortcode)

	return err
}

// handle for GET and POST
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("=> " + r.Method + " " + r.URL.Path)

	// handle GET request
	if r.Method == "GET" {
		data := DefaultResponse{
			Message: "🧑🏽‍💻 Welcome to URL shortener",
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Write(jsonData)

		return
	}

	// handle POST request
	if r.Method == "POST" {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var payload Payload
		err = json.Unmarshal(bodyBytes, &payload)
		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusInternalServerError)
			return
		}

		// set default values if missing
		if payload.Single == nil {
			defaultSingle := false
			payload.Single = &defaultSingle
		}
		if payload.Expires == nil {
			// defaults to 1000 years
			defaultExpires := int64(1000) * 365 * 24 * 60 * 60
			payload.Expires = &defaultExpires
		}

		// generate a shortcode
		// for simplicity, encode current timestamp
		shortcode := encodeToBase62(time.Now().UnixNano())

		// save the URL to the database
		err = saveURLToDB(payload, shortcode)
		if err != nil {
			http.Error(w, "Error saving to database", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"shortcode":"%s"}`, shortcode)))

		return
	}
}

// redirect handler for the shortcode
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortcode := r.URL.Path[len("/short/"):]

	log.Println("=> " + r.Method + " " + r.URL.Path + " " + shortcode)

	// get the URL from the database
	url, singleUse, expiresAt, err := getURLFromDB(shortcode)
	if err != nil || url == "" {
		http.Error(w, "No URL found for this shortcode", http.StatusNotFound)
		return
	}

	// check if the link has expired
	if time.Now().After(expiresAt) {
		// delete expired link
		deleteURLFromDB(shortcode)
		http.Error(w, "URL for this shortcode has expired", http.StatusGone)
		return
	}

	// redirect to the original URL
	http.Redirect(w, r, url, http.StatusFound)

	// if it's a single-use link, delete it after redirecting
	if singleUse {
		deleteURLFromDB(shortcode)
	}
}

// clean expired links
func cleanExpiredLinks() {
	_, err := dbPool.Exec(context.Background(), "DELETE FROM urls WHERE expires_at < $1", time.Now())

	if err != nil {
		log.Println("😭 Error cleaning expired links:", err)
	}
}

func main() {
	// check if PLATFORM_APPLICATION environment variable is defined
	if os.Getenv("PLATFORM_APPLICATION") != "" {
		// initialize blackfire profiler
		p_err := profiler.Start(
			profiler.WithAppName("url-shortener-golang"),
		)
		if p_err != nil {
			panic("😭 Error while starting Blackfire profiler")
		}

		defer profiler.Stop()
		log.Println("👾 Blackfire profiler started")
	} else {
		log.Println("🦦 PLATFORM_APPLICATION not set. Skipping Blackfire profiler initialization.")
	}

	// connect to PostgreSQL
	connectToDB()
	defer dbPool.Close()

	// create URLs table if it doesn't exist
	createTableIfNotExists()

	// get port from env
	port := os.Getenv("PORT")
	if port == "" {
		// default port
		port = "1001"
	}

	// set up routes
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/short/", redirectHandler)

	// set up cron job to clean expired links
	c := cron.New()
	c.AddFunc("@every 5m", cleanExpiredLinks)
	c.Start()

	// start server
	log.Println("🤖 Server started on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("❌ Unable to start server: %v", err)
	}
}
