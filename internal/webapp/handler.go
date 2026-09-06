package webapp

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/ashwingopalsamy/llm-txt-trimmer/internal/compact"
)

const maxRequestBytes = 4 << 20

//go:embed static/*
var staticFiles embed.FS

type trimRequest struct {
	Text string       `json:"text"`
	Mode compact.Mode `json:"mode"`
}

type trimStats struct {
	BeforeBytes      int     `json:"beforeBytes"`
	AfterBytes       int     `json:"afterBytes"`
	SavedBytes       int     `json:"savedBytes"`
	SavedPercent     float64 `json:"savedPercent"`
	BeforeWhitespace int     `json:"beforeWhitespace"`
	AfterWhitespace  int     `json:"afterWhitespace"`
	BeforeLines      int     `json:"beforeLines"`
	AfterLines       int     `json:"afterLines"`
	TokenProxyBefore int     `json:"tokenProxyBefore"`
	TokenProxyAfter  int     `json:"tokenProxyAfter"`
}

type trimResponse struct {
	Text  string    `json:"text"`
	Stats trimStats `json:"stats"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/trim", handleTrim)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(staticRoot))
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		fileServer.ServeHTTP(w, r)
	}))

	return securityHeaders(mux)
}

func handleTrim(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	defer r.Body.Close()

	var req trimRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "input is too large; maximum request size is 4 MiB")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	out, stats, err := compact.Transform(req.Text, req.Mode)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := trimResponse{
		Text: out,
		Stats: trimStats{
			BeforeBytes:      stats.BeforeBytes,
			AfterBytes:       stats.AfterBytes,
			SavedBytes:       stats.SavedBytes(),
			SavedPercent:     stats.SavedPercent(),
			BeforeWhitespace: stats.BeforeWhitespace,
			AfterWhitespace:  stats.AfterWhitespace,
			BeforeLines:      stats.BeforeLines,
			AfterLines:       stats.AfterLines,
			TokenProxyBefore: stats.TokenProxyBefore(),
			TokenProxyAfter:  stats.TokenProxyAfter(),
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
