package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHaversineKm_ZeroDistance(t *testing.T) {
	d := haversineKm(0, 0, 0, 0)
	if d != 0 {
		t.Fatalf("expected 0km, got %v", d)
	}
}

func TestParseLatLngRadius_ValidDefaults(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/any?lat=10&lng=20", nil)
	c.Request = req

	lat, lng, radius, ok := parseLatLngRadius(c)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if lat != 10 || lng != 20 {
		t.Fatalf("unexpected lat/lng: %v,%v", lat, lng)
	}
	if radius != 25 {
		t.Fatalf("expected default radius 25km, got %v", radius)
	}
}

func TestParseLatLngRadius_MissingLat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/any?lng=20", nil)
	c.Request = req

	_, _, _, ok := parseLatLngRadius(c)
	if ok {
		t.Fatalf("expected ok=false for missing lat")
	}
}

func TestParseLatLngRadius_OutOfBounds(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("GET", "/any?lat=100&lng=20", nil)
	c.Request = req

	_, _, _, ok := parseLatLngRadius(c)
	if ok {
		t.Fatalf("expected ok=false for out-of-bounds lat")
	}
}

