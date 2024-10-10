package main

import (
	"fmt"
    "os"
    "log"
	"math/rand"
	"net/http"
	"strings"
	"time"
    "runtime"
)

const timeFormat = time.RFC3339

type logWriter struct{}

func (lw *logWriter) Write(bs []byte) (int, error) {
    return fmt.Print(time.Now().UTC().Format(timeFormat), " | ", string(bs))
}

var (
    longUrls = make(map[string]string)
    urls = make(map[string]string)
    InfoLog *log.Logger
    ErrorLog *log.Logger
)

func main() {
	http.HandleFunc("/", handleForm)
	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/short/", handleRedirect)

    InfoLog = log.New(os.Stdout, "INFO: ", 0)
    ErrorLog = log.New(os.Stdout, "ERROR: ", 0)
    InfoLog.SetOutput(new(logWriter))
    ErrorLog.SetOutput(new(logWriter))

	InfoLog.Println("URL Shortener is running on :3030")
	http.ListenAndServe(":3030", nil)
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/shorten", http.StatusSeeOther)
		return
	}

	// Serve the HTML form
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<form method="post" action="/shorten">
				<input type="url" name="url" placeholder="Enter a URL" required>
				<input type="submit" value="Shorten">
			</form>
		</body>
		</html>
	`)
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
    InfoLog.Println("function: ", CallerName(0))
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        ErrorLog.Println(http.StatusMethodNotAllowed, "Invalid request method", r.Method)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
        ErrorLog.Println(http.StatusBadRequest, "URL parameter is missing")
		return
	}

    InfoLog.Println("original URL: ", originalURL)

	// Generate a unique shortened key for the original URL
	shortKey := generateShortKey(originalURL)

    _, ok := longUrls[originalURL]
    if ok {
        InfoLog.Println("Repeated URL generation")
    }

	// Construct the full shortened URL
	shortenedURL := fmt.Sprintf("http://localhost:3030/short/%s", shortKey)

    if !ok {
        longUrls[originalURL] = shortKey
	    urls[shortKey] = originalURL
        InfoLog.Println("generated URL: ", shortenedURL)
    }

	// Serve the result page
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<p>Original URL: `, originalURL, `</p>
			<p>Shortened URL: <a href="`, shortenedURL, `">`, shortenedURL, `</a></p>
		</body>
		</html>
	`)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
    InfoLog.Println("function: ", CallerName(0))
	shortKey := strings.TrimPrefix(r.URL.Path, "/short/")
	if shortKey == "" {
		http.Error(w, "Shortened key is missing", http.StatusBadRequest)
        ErrorLog.Println(http.StatusBadRequest, "Shortened key is missing")
		return
	}

    InfoLog.Println("short key: ", shortKey)

	// Retrieve the original URL from the `urls` map using the shortened key
	originalURL, found := urls[shortKey]
	if !found {
		http.Error(w, "Shortened key not found", http.StatusNotFound)
        ErrorLog.Println(http.StatusNotFound, "Shortened key is missing")
		return
	}

    InfoLog.Println("redirecting to URL: ", originalURL)

	// Redirect the user to the original URL
	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func generateShortKey(url string) string {
    InfoLog.Println("function: ", CallerName(0))
    InfoLog.Println("called from: ", CallerName(1))

    val, ok := longUrls[url]
    if ok {
        InfoLog.Println("short key already exists")
        InfoLog.Println("returning", val)
        return val
    }

	const charset = "abcdefghijklmnopqrstuvxyzABCDEFGHIJKLMNOPQRSTUVXYZ0123456789"
	const keyLength = 6

	rand.Seed(time.Now().UnixNano())
	shortKey := make([]byte, keyLength)

	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}

    InfoLog.Println("generated new short Key: ", string(shortKey))

	return string(shortKey)
}

func CallerName(skip int) string {
        pc, _, _, ok := runtime.Caller(skip + 1)
        if !ok {
                return ""
        }
        f := runtime.FuncForPC(pc)
        if f == nil {
                return ""
        }
        return f.Name()
}
