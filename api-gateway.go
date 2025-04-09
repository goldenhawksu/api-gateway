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

const (
	httpbinProxyPath = "/get"
	httpbinTarget    = "https://httpbin.org/get"
)

var deniedHeaders = []string{"host", "referer", "cf-", "forward", "cdn"}

func isAllowedHeader(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, deniedHeader := range deniedHeaders {
		if strings.Contains(lowerKey, deniedHeader) {
			return false
		}
	}
	return true
}

func targetURL(pathname string) string {
	split := strings.Index(pathname[1:], "/")
	if split == -1 {
		return ""
	}
	prefix := pathname[:split+2] // +2 包含斜杠
	if base, exists := apiMapping[prefix]; exists {
		return base + pathname[len(prefix):]
	}
	return ""
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Handle root path
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Service is running!")
		return
	}

	// Handle robots.txt
	if r.URL.Path == "/robots.txt" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "User-agent: *\nDisallow: /")
		return
	}

	// Handle httpbin proxy
	if r.URL.Path == httpbinProxyPath {
		handleHttpbinProxy(w, r)
		return
	}

	// Original proxy logic
	query := r.URL.RawQuery
	if query != "" {
		query = "?" + query
	}

	targetURL := targetURL(r.URL.Path + query)
	if targetURL == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	proxyRequest(w, r, targetURL)
}

func handleHttpbinProxy(w http.ResponseWriter, r *http.Request) {
	query := r.URL.RawQuery
	target := httpbinTarget
	if query != "" {
		target += "?" + query
	}
	proxyRequest(w, r, target)
}

// 通用代理逻辑封装
func proxyRequest(w http.ResponseWriter, r *http.Request, target string) {
	client := &http.Client{}
	proxyReq, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		if isAllowedHeader(key) {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to fetch: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 安全头设置
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")

	w.WriteHeader(resp.StatusCode)
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
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}