package handlers

import (
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
	httpClient *http.Client
}

func NewGeocodeHandler(baseURL, userAgent string) *GeocodeHandler {
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
