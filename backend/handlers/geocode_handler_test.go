package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestGeocodeSearch_MissingQ(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/geocode", nil)
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "patient@healthonyx.demo",
		Role:   "patient",
	})

	h := NewGeocodeHandler("", "")
	h.GeocodeSearch(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGeocodeSearch_ParsesUpstreamResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/search" {
			t.Fatalf("expected path /search, got %s", req.URL.Path)
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`[
			{"lat":"29.6516","lon":"-82.3248","display_name":"Gainesville, FL"},
			{"lat":"bad","lon":"-82.3248","display_name":"Invalid row"}
		]`))
	}))
	defer ts.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/geocode?q=Gainesville&limit=2", nil)
	c.Set("claims", &Claims{
		UserID: uuid.New(),
		Email:  "patient@healthonyx.demo",
		Role:   "patient",
	})

	h := NewGeocodeHandler(ts.URL, "Healthonyx-test/1.0")
	h.GeocodeSearch(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
	var payload struct {
		Results []struct {
			Lat         float64 `json:"lat"`
			Lng         float64 `json:"lng"`
			DisplayName string  `json:"display_name"`
		} `json:"results"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected valid JSON, got error: %v, body: %s", err, w.Body.String())
	}
	if len(payload.Results) != 1 {
		t.Fatalf("expected 1 parsed result (invalid row filtered), got %d", len(payload.Results))
	}
	if payload.Results[0].DisplayName != "Gainesville, FL" {
		t.Fatalf("unexpected display_name: %s", payload.Results[0].DisplayName)
	}
}

