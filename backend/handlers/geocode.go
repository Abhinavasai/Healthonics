package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GeocodeHandler proxies search requests to OpenStreetMap Nominatim so the browser
// avoids CORS issues and sends a proper User-Agent per Nominatim usage policy.
// Next sprint you can swap the upstream URL to a self-hosted Nominatim or add
// Google Geocoding behind the same route shape.
type GeocodeHandler struct {
	baseURL    string
	userAgent  string
	googleKey  string
	httpClient *http.Client
}

func NewGeocodeHandler(baseURL, userAgent, googleKey string) *GeocodeHandler {
	b := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if b == "" {
		b = "https://nominatim.openstreetmap.org"
	}
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		ua = "Healthonyx/1.0 (https://github.com/Abhinavasai/Healthonyx)"
	}
	return &GeocodeHandler{
		baseURL:   b,
		userAgent: ua,
		googleKey: strings.TrimSpace(googleKey),
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

type geocodeHitJSON struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	DisplayName string  `json:"display_name"`
}

type nominatimRow struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

type googleGeocodeRow struct {
	FormattedAddress string `json:"formatted_address"`
	Geometry         struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
}

type googleGeocodeResponse struct {
	Results []googleGeocodeRow `json:"results"`
	Status  string             `json:"status"`
}

type nearbyHospitalsRequest struct {
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	RadiusKM   float64 `json:"radius_km"`
	Department string  `json:"department"`
}

type googleNearbyPlacesRequest struct {
	IncludedTypes       []string `json:"includedTypes,omitempty"`
	MaxResultCount      int      `json:"maxResultCount,omitempty"`
	LocationRestriction struct {
		Circle struct {
			Center struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"center"`
			Radius float64 `json:"radius"`
		} `json:"circle"`
	} `json:"locationRestriction"`
	TextQuery string `json:"textQuery,omitempty"`
}

type googleNearbyPlace struct {
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	FormattedAddress string `json:"formattedAddress"`
	Location         struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
}

type googleNearbyPlacesResponse struct {
	Places []googleNearbyPlace `json:"places"`
}

// GeocodeSearch handles GET /api/geocode?q=...&limit=5
func (h *GeocodeHandler) GeocodeSearch(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q query parameter is required"})
		return
	}

	limit := 5
	if ls := strings.TrimSpace(c.Query("limit")); ls != "" {
		n, err := strconv.Atoi(ls)
		if err != nil || n < 1 || n > 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 10"})
			return
		}
		limit = n
	}

	if h.googleKey != "" {
		out, err := h.googleGeocode(c.Request.Context(), q, limit)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"results": out})
			return
		}
	}

	u, err := url.Parse(h.baseURL + "/search")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "geocode configuration error"})
		return
	}
	qv := url.Values{}
	qv.Set("q", q)
	qv.Set("format", "json")
	qv.Set("limit", fmt.Sprintf("%d", limit))
	qv.Set("addressdetails", "0")
	u.RawQuery = qv.Encode()

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, u.String(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "geocode request error"})
		return
	}
	req.Header.Set("User-Agent", h.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "geocode service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "geocode read error"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"error": "geocode upstream error"})
		return
	}

	var rows []nominatimRow
	if err := json.Unmarshal(body, &rows); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "geocode parse error"})
		return
	}

	out := make([]geocodeHitJSON, 0, len(rows))
	for _, r := range rows {
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(r.Lat), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(r.Lon), 64)
		if err1 != nil || err2 != nil || lat < -90 || lat > 90 || lng < -180 || lng > 180 {
			continue
		}
		out = append(out, geocodeHitJSON{
			Lat:         lat,
			Lng:         lng,
			DisplayName: r.DisplayName,
		})
	}

	c.JSON(http.StatusOK, gin.H{"results": out})
}

func (h *GeocodeHandler) googleGeocode(ctx context.Context, query string, limit int) ([]geocodeHitJSON, error) {
	u, err := url.Parse("https://maps.googleapis.com/maps/api/geocode/json")
	if err != nil {
		return nil, err
	}
	qv := url.Values{}
	qv.Set("address", query)
	qv.Set("key", h.googleKey)
	u.RawQuery = qv.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google geocode status=%d", resp.StatusCode)
	}
	var parsed googleGeocodeResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	out := make([]geocodeHitJSON, 0, min(limit, len(parsed.Results)))
	for i, row := range parsed.Results {
		if i >= limit {
			break
		}
		out = append(out, geocodeHitJSON{
			Lat:         row.Geometry.Location.Lat,
			Lng:         row.Geometry.Location.Lng,
			DisplayName: row.FormattedAddress,
		})
	}
	return out, nil
}

// NearbyHospitalsGoogle handles POST /api/find-care/places/nearby.
func (h *GeocodeHandler) NearbyHospitalsGoogle(c *gin.Context) {
	if _, ok := getClaims(c); !ok {
		return
	}
	if h.googleKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Maps API key not configured"})
		return
	}
	var reqBody nearbyHospitalsRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if reqBody.Lat < -90 || reqBody.Lat > 90 || reqBody.Lng < -180 || reqBody.Lng > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coordinates"})
		return
	}
	if reqBody.RadiusKM <= 0 {
		reqBody.RadiusKM = 25
	}
	if reqBody.RadiusKM > 200 {
		reqBody.RadiusKM = 200
	}

	payload := googleNearbyPlacesRequest{
		IncludedTypes:  []string{"hospital"},
		MaxResultCount: 15,
	}
	payload.LocationRestriction.Circle.Center.Latitude = reqBody.Lat
	payload.LocationRestriction.Circle.Center.Longitude = reqBody.Lng
	payload.LocationRestriction.Circle.Radius = reqBody.RadiusKM * 1000
	if q := strings.TrimSpace(reqBody.Department); q != "" {
		payload.TextQuery = q + " hospital"
	}
	rawBody, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodPost,
		"https://places.googleapis.com/v1/places:searchNearby",
		bytes.NewReader(rawBody),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not build places request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Goog-Api-Key", h.googleKey)
	httpReq.Header.Set("X-Goog-FieldMask", "places.displayName,places.formattedAddress,places.location")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Google Places request failed"})
		return
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Could not read Google Places response"})
		return
	}
	if resp.StatusCode >= http.StatusBadRequest {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Google Places search failed"})
		return
	}
	var parsed googleNearbyPlacesResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Could not parse Google Places response"})
		return
	}

	type placeRow struct {
		Name      string  `json:"name"`
		Address   string  `json:"address"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	out := make([]placeRow, 0, len(parsed.Places))
	for _, p := range parsed.Places {
		out = append(out, placeRow{
			Name:      p.DisplayName.Text,
			Address:   p.FormattedAddress,
			Latitude:  p.Location.Latitude,
			Longitude: p.Location.Longitude,
		})
	}
	c.JSON(http.StatusOK, gin.H{"hospitals": out})
}
