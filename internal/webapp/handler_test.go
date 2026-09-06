package webapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTrimAPI(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/trim", strings.NewReader(`{"text":"hello   world\nagain","mode":"dense"}`))
	rec := httptest.NewRecorder()

	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Text != "hello world again" {
		t.Fatalf("got %q", body.Text)
	}
}

func TestTrimAPIRejectsInvalidMode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/trim", strings.NewReader(`{"text":"x","mode":"nope"}`))
	rec := httptest.NewRecorder()

	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestServesIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "llmtrim") {
		t.Fatalf("index does not contain app name")
	}
}
