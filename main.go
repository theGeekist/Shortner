package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv" // for loading .env files
	"go.uber.org/zap"          // for structured logging
	"go.uber.org/zap/zapcore"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db     *sql.DB
	Logger *zap.SugaredLogger
)

// trackingParams holds common query parameters to remove.
var trackingParams = []string{
	"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
	"fbclid", "gclid", "mc_eid", "mc_cid", "trk", "msclkid", "dclid",
}

// fullyDecode repeatedly decodes a URL-encoded string until no changes occur.
func fullyDecode(raw string) string {
	decoded := raw
	for {
		temp, err := url.QueryUnescape(decoded)
		if err != nil || temp == decoded {
			break
		}
		decoded = temp
	}
	return decoded
}

// cleanURL performs several cleaning steps on a raw URL string:
// - Trims whitespace and extraneous angle brackets (common in email clients)
// - Fully decodes URL encoding (to handle double-encoded URLs)
// - Removes tracking parameters
// - If the URL is wrapped in a Google/Outlook redirect, extracts the real URL.
func cleanURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.Trim(rawURL, "<>")

	decodedURL := fullyDecode(rawURL)

	parsedURL, err := url.Parse(decodedURL)
	if err != nil {
		return "", err
	}

	// Remove tracking parameters.
	q := parsedURL.Query()
	for _, param := range trackingParams {
		q.Del(param)
	}
	parsedURL.RawQuery = q.Encode()

	// Check if the URL is wrapped in a Google/Outlook redirect and extract the real URL.
	re := regexp.MustCompile(`https?://(www\.)?(l|out)\.google\.com/url\?.*?url=([^&]+)`)
	matches := re.FindStringSubmatch(parsedURL.String())
	if len(matches) == 4 {
		extracted, err := url.QueryUnescape(matches[3])
		if err == nil {
			return extracted, nil
		}
	}

	return parsedURL.String(), nil
}

// generateCode creates a random alphanumeric string of length n.
func generateCode(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// shortenHandler processes incoming URL-shortening requests.
func shortenHandler(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	cleanedURL, err := cleanURL(rawURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	code := generateCode(6)
	_, err = db.Exec("INSERT INTO links (code, url, created_at) VALUES (?, ?, ?)", code, cleanedURL, time.Now().UTC())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Use the domain from the environment.
	domain := os.Getenv("SHORT_DOMAIN")
	if domain == "" {
		domain = "http://localhost"
	}
	fmt.Fprintf(w, "%s/%s", domain, code)
}

// redirectHandler looks up the original URL from the code and redirects the client.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	if code == "" {
		http.NotFound(w, r)
		return
	}
	var urlStr string
	err := db.QueryRow("SELECT url FROM links WHERE code = ?", code).Scan(&urlStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, urlStr, http.StatusFound)
}

// cleanupOldLinks deletes entries older than CACHE_EXPIRY days.
func cleanupOldLinks() {
	expiry := os.Getenv("CACHE_EXPIRY")
	if expiry == "" {
		expiry = "30"
	}
	query := fmt.Sprintf("DELETE FROM links WHERE created_at < datetime('now', '-%s days')", expiry)
	_, err := db.Exec(query)
	if err != nil {
		Logger.Errorf("Error cleaning up old links: %v", err)
	} else {
		Logger.Infoln("Old links cleaned up successfully")
	}
}

// initLogger initialises the zap logger based on environment variables.
func initLogger() {
	// Get log file path and level from environment; use defaults if not set.
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "./shortner.log"
	}
	logLevelStr := os.Getenv("LOG_LEVEL")
	var logLevel zapcore.Level
	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = zap.DebugLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "info":
		logLevel = zap.InfoLevel
	default:
		logLevel = zap.InfoLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(logLevel),
		Development:      false, // Adjust as needed (e.g., set to true if in development mode)
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout", logPath},
		ErrorOutputPaths: []string{"stderr"},
	}
	l, err := cfg.Build()
	if err != nil {
		fmt.Printf("can't initialize zap logger: %v\n", err)
		os.Exit(1)
	}
	Logger = l.Sugar()
}

func main() {
	// Load environment variables from .env file (if present).
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}
	// Initialise zap logger.
	initLogger()
	Logger.Infoln("Logger initialised.")

	rand.Seed(time.Now().UnixNano())

	var err error
	db, err = sql.Open("sqlite3", "./shortlinks.db")
	if err != nil {
		Logger.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create the links table if it doesn't exist.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE,
		url TEXT,
		created_at TIMESTAMP
	);`)
	if err != nil {
		Logger.Fatalf("Error creating table: %v", err)
	}

	// Schedule a daily cleanup of links older than CACHE_EXPIRY days.
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for {
			<-ticker.C
			cleanupOldLinks()
		}
	}()

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	// Read the server port from environment.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8889"
	}
	Logger.Infoln("Starting HTTP server on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		Logger.Fatalf("Server failed: %v", err)
	}
}

