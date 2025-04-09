package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var apiMapping = map[string]string{
	"/discord":     "https://discord.com/api",
	"/telegram":    "https://api.telegram.org",
	"/openai":      "https://api.openai.com",
	"/claude":      "https://api.anthropic.com",
	"/gemini":      "https://generativelanguage.googleapis.com",
	"/meta":        "https://www.meta.ai/api",
	"/groq":        "https://api.groq.com/openai",
	"/xai":         "https://api.x.ai",
	"/cohere":      "https://api.cohere.ai",
	"/huggingface": "https://api-inference.huggingface.co",
	"/together":    "https://api.together.xyz",
	"/novita":      "https://api.novita.ai",
	"/portkey":     "https://api.portkey.ai",
	"/fireworks":   "https://api.fireworks.ai",
	"/openrouter":  "https://openrouter.ai/api",
	"/cerebras":    "https://api.cerebras.ai",
}

var deniedHeaders = []string{"host", "referer", "cf-", "forward", "cdn"}

func isAllowedHeader(key string) bool {
	for _, deniedHeader := range deniedHeaders {
		if strings.Contains(strings.ToLower(key), deniedHeader) {
			return false
		}
	}
	return true
}

func targetURL(pathname string) string {
	split := strings.Index(pathname[1:], "/")
	prefix := pathname[:split+1]
	if base, exists := apiMapping[prefix]; exists {
		return base + pathname[len(prefix):]
	}
	return ""
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Service is running!")
		return
	}

	if r.URL.Path == "/robots.txt" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "User-agent: *\nDisallow: /")
		return
	}

	query := r.URL.RawQuery

	if query != "" {
		query = "?" + query
	}

	targetURL := targetURL(r.URL.Path + query)

	if targetURL == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Create new request
	client := &http.Client{}
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		if isAllowedHeader(key) {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Make the request
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to fetch: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response: %v", err)
	}
}

func main() {
	port := "2233"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	http.HandleFunc("/", handler)
	log.Printf("Starting server on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}