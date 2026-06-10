package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const upstream = "https://openrouter.ai/api/v1/chat/completions"

var logDir string

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Println("warning: OPENROUTER_API_KEY not set — forwarding requests without Authorization header")
	}

	logDir = os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "/logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("failed to create log dir %s: %v", logDir, err)
	}
	log.Printf("writing request/response logs to %s", logDir)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		handleChatCompletion(w, r, apiKey)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	addr := ":" + port
	log.Printf("openrouter-proxy listening on %s → %s", addr, upstream)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// logEntry is the structure written to each JSON log file.
type logEntry struct {
	Timestamp string          `json:"timestamp"`
	Request   json.RawMessage `json:"request"`
	Response  json.RawMessage `json:"response,omitempty"`
	Error     string          `json:"error,omitempty"`
}

func writeLog(ts time.Time, entry logEntry) {
	filename := fmt.Sprintf("%s/%s.json", logDir, ts.UTC().Format("20060102T150405.000000000Z"))
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		log.Printf("log marshal error: %v", err)
		return
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("log write error: %v", err)
	}
}

func handleChatCompletion(w http.ResponseWriter, r *http.Request, apiKey string) {
	ts := time.Now()

	// Buffer request body so we can log it and forward it
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	// Build upstream request
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstream, bytes.NewReader(reqBody))
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}

	// Forward relevant request headers; replace Authorization only if key is set
	for key, vals := range r.Header {
		switch http.CanonicalHeaderKey(key) {
		case "Authorization", "Content-Length":
			// Skip: Authorization replaced below (if key set); Content-Length let Go recalculate
		default:
			for _, v := range vals {
				req.Header.Add(key, v)
			}
		}
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeLog(ts, logEntry{
			Timestamp: ts.UTC().Format(time.RFC3339Nano),
			Request:   json.RawMessage(reqBody),
			Error:     err.Error(),
		})
		http.Error(w, "upstream request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Buffer response body so we can log it and forward it
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("upstream read error: %v", err)
	}

	// Represent response as JSON: inline if valid JSON, string otherwise
	var respJSON json.RawMessage
	if json.Valid(respBody) {
		respJSON = json.RawMessage(respBody)
	} else {
		quoted, _ := json.Marshal(string(respBody))
		respJSON = json.RawMessage(quoted)
	}
	writeLog(ts, logEntry{
		Timestamp: ts.UTC().Format(time.RFC3339Nano),
		Request:   json.RawMessage(reqBody),
		Response:  respJSON,
	})

	// Forward all response headers verbatim, except Content-Length
	for key, vals := range resp.Header {
		if http.CanonicalHeaderKey(key) == "Content-Length" {
			continue
		}
		for _, v := range vals {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}
