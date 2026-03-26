# Sprint 2 – Submission Summary

**Team:** Healthonyx AI Collective  
**Sprint:** Sprint 2 (Integrated Geo Find-Care Flow + Quality Gates)  
**Repository:** https://github.com/Abhinavasai/Healthonyx  
**Sprint 2 branch:** `dev` (updates merged here)

---

## 1. Sprint 1 Review (and what Sprint 2 builds on)

In **Sprint 1**, the project delivered authentication + role-based routing and placeholder dashboards across:
- Backend: DB schema + migrations, JWT auth, RBAC
- Frontend: Register/Login, AuthGuard/RoleGuard, AppShell with logout

Sprint 1 documentation (`Sprint1.md`) indicates all *planned Sprint 1 stories* were completed and integrated into `dev`.

### Sprint 2 goal
Add an integrated “Find care near you” geo flow that works like real-world delivery-app UX:
- User searches a location (text/address)
- Or uses browser location access (“Use my location”)
- App shows **nearby hospitals** (free map + distance filtering)
- Foundation is laid to plug in a richer maps/geocoding provider (Google Maps later), while keeping the same frontend/backend contract.

---

## 2. What We Completed in Sprint 2 (Detailed Progress)

### 2.1 Integrate frontend + backend

#### Backend: new geocode + geo search API
We added a backend route that converts **text location → coordinates**:
- `GET /api/geocode?q=...&limit=...` (patient role protected)
  - Server proxies requests to **OpenStreetMap Nominatim** (free service) to avoid browser CORS issues.
  - Returns: `{ results: [{ lat, lng, display_name }, ...] }`

We also built on the existing geo-search endpoints:
- `GET /api/hospitals/near?lat=...&lng=...&radius_km=...` (patient role protected)
  - Uses `haversineKm` server-side to return hospitals within the radius

#### Frontend: “Find care near you” UI is connected
On the patient side we implemented/updated the route:
- `/patient/find-care` → `PatientFindCareComponent`

The page now includes:
- Location **search bar** (text)
- “Use my location” button using `navigator.geolocation`
- A **free map** (Leaflet + OpenStreetMap tiles)
- After selecting/searching a location:
  - Hospitals can be loaded via `Hospitals nearby`
  - Hospitals are displayed both as a list and as map markers

### 2.2 Free map options (non-Google)
Sprint 2 uses:
- OpenStreetMap tiles (free)
- Leaflet (free)
- Nominatim (free, with strict usage policies)

This makes the app feel realistic today, and also creates a clean path to integrate Google Maps in Sprint 3 without rewriting the hospital-distance backend contract.

---

## 3. Backend API Documentation (Sprint 2)

### Authentication (role-based)
All API routes below require a valid JWT header where noted.

#### `GET /health`
- **Auth:** None
- **Response (200):**
  - `{ "status": "ok" }`

#### `POST /api/register`
- **Auth:** None
- **Body:**
  - `email: string`
  - `password: string` (min 6 chars)
  - `role: "patient" | "doctor" | "admin"`
- **Response (201):**
  - `{ "token": "<jwt>", "user": { "id": "<uuid>", "email": "...", "role": "..." } }`
- **Errors:**
  - `400` invalid request/role/password/email
  - `409` email already registered

#### `POST /api/login`
- **Auth:** None
- **Body:**
  - `email: string`
  - `password: string`
- **Response (200):**
  - `{ "token": "<jwt>", "user": { "id": "<uuid>", "email": "...", "role": "..." } }`
- **Errors:**
  - `401` invalid email/password

#### `GET /api/me`
- **Auth:** Required (`Authorization: Bearer <token>`)
- **Response (200):**
  - `{ "id": "<uuid>", "email": "...", "role": "..." }`

### Role-protected placeholders
These are demonstration endpoints that show RBAC behavior:
- `GET /api/patient` (patient role)
- `GET /api/doctor` (doctor role)
- `GET /api/admin` (admin role)

---

## 4. Geo / Hospitals API (Sprint 2)

All endpoints below are **patient role protected**.

### Location search (text → coordinates)
#### `GET /api/geocode?q=<query>&limit=<n>`
- **Auth:** Patient required
- **Query params:**
  - `q` (required) location text (city/address/landmark)
  - `limit` (optional, default 5, range 1–10)
- **Response (200):**
  - `{ "results": [ { "lat": number, "lng": number, "display_name": string }, ... ] }`
- **Errors:**
  - `400` when `q` missing or `limit` invalid
  - `502` when upstream geocoding fails

### Nearby hospitals
#### `GET /api/hospitals/near?lat=<lat>&lng=<lng>&radius_km=<radius>`
- **Auth:** Patient required
- **Query params:**
  - `lat` required; numeric; valid range `[-90, 90]`
  - `lng` required; numeric; valid range `[-180, 180]`
  - `radius_km` optional; default `25`; must be `> 0` and `<= 500`
- **Response (200):**
  - `{ "hospitals": [ { "id": string, "name": string, "city": string, "region": string, "latitude": number, "longitude": number, "distance_km": number }, ... ] }`
- **Errors:**
  - `400` for invalid/missing lat/lng/radius
  - `500` internal errors

### Nearby doctors (supports specialty keyword filter)
#### `GET /api/doctors/search?lat=<lat>&lng=<lng>&radius_km=<radius>&specialization=<keyword>`
- **Auth:** Patient required
- **Query params:**
  - Same `lat`, `lng`, `radius_km` rules as hospitals
  - `specialization` optional; substring match (case-insensitive)
- **Response (200):**
  - `{ "doctors": [ { "id": string, "email": string, "specialization": string, "hospital_id": string|null, "distance_km": number }, ... ] }`

---

## 5. Appointment Availability / Booking API (used by the geo booking flow)

These are already present in `GeoBookingHandler` and can be connected from the new find-care UX.

### Slots list
#### `GET /api/doctors/:id/slots`
- **Auth:** Patient required
- **Response (200):**
  - `{ "slots": [ { "id": string, "doctor_id": string, "start_at": "<iso>", "end_at": "<iso>", "available": boolean }, ... ] }`

### Doctor creates a slot
#### `POST /api/doctor/slots`
- **Auth:** Doctor required
- **Body:**
  - `start_at`: time
  - `end_at`: time (must be after `start_at`)
- **Response (201):**
  - `{ "id": "<uuid>" }`

### Doctor deletes an open slot
#### `DELETE /api/doctor/slots/:id`
- **Auth:** Doctor required
- **Response (200):**
  - `{ "ok": true }`

### Patient books a slot
#### `POST /api/appointments/book-slot`
- **Auth:** Patient required
- **Body:**
  - `slot_id`: string (UUID)
  - `reason`: string
- **Response (201):**
  - Returns an appointment object (same shape as appointments create)
- **Errors:**
  - `400` invalid body/UUID
  - `404` slot not found
  - `409` slot already booked / no longer available

---

## 6. Sprint 2 Testing (Unit + Cypress)

### 6.1 Backend unit tests (Go)
We added backend unit tests for geo parsing/math helpers and the new geocoding handler.

#### Test files added
- `backend/handlers/geo_booking_test.go`
  - `TestHaversineKm_ZeroDistance`
  - `TestParseLatLngRadius_ValidDefaults`
  - `TestParseLatLngRadius_MissingLat`
  - `TestParseLatLngRadius_OutOfBounds`
- `backend/handlers/geocode_handler_test.go`
  - `TestGeocodeSearch_MissingQ`
  - `TestGeocodeSearch_ParsesUpstreamResults` (uses `httptest` to stub upstream Nominatim `/search`)

**Command run:**
```powershell
cd backend
go test ./...
```
**Result:** Passed (all handler/helper tests succeeded).

### 6.2 Frontend unit tests (Angular)
We added unit tests for both the geo service layer and the patient “find care” component UI/state logic.

#### Test files added
- `frontend/src/app/services/geo-booking.service.spec.ts`
  - tests `geocodeSearch` request params + response shape
  - tests `hospitalsNear` request params + response handling
- `frontend/src/app/components/patient-find-care/patient-find-care.component.spec.ts`
  - verifies component error handling when `locationQuery` is empty
  - verifies successful `runGeocodeSearch()` updates `lat/lng` + selected label (single-hit path)
  - verifies `pickGeocodeSuggestion()` updates location + coordinates
  - verifies `useMyLocation()` permission-denied error path (mocked via Jasmine `spyOnProperty`)
  - verifies `loadHospitals()` calls the correct API arguments and updates `hospitals`
  - verifies `loadDoctors()` calls the correct API arguments and updates `doctors`

**Command run:**
```powershell
cd frontend
CI=true npx ng test --watch=false --browsers=ChromeHeadless --progress=false
```
**Result:** Passed (all Angular unit tests succeeded).

### 6.3 Cypress e2e test (very simple integration)
We added a minimal e2e test suite for the patient find-care page.

#### Cypress files
- `frontend/cypress/e2e/find-care.cy.js`
  - stubs `/api/geocode` (single hit)
  - stubs `/api/hospitals/near` and verifies the hospital list renders
  - stubs `/api/doctors/search` and verifies the doctors list renders

The Cypress test uses `data-cy` selectors added to `PatientFindCareComponent` to keep selectors stable.

**Command run:**
```powershell
cd frontend
npx cypress run --headless --spec "cypress/e2e/find-care.cy.js"
```
**Result:** Passed (1 spec, 2 tests passing).

---

## 7. Improvements for Sprint 3 (Suggested Next Steps)

1. **Map markers for doctors** (not only hospitals)
2. **Booking flow UI** from the find-care page:
   - show available slots per doctor
   - allow patient to book a slot (calls existing slot/booking endpoints)
3. **Performance improvements for geo search**:
   - add DB-side filtering or PostGIS for large hospital datasets
   - avoid loading all hospitals for every request
4. **Geocode caching** (store recent queries to reduce external calls and speed up UX)
5. **Google Maps integration path**:
   - keep the `/api/geocode` response contract so the frontend can swap providers easily
   - add a real map UI and autocomplete while staying aligned with Sprint 2 API contracts

---

## 8. Video / Narration Plan (Entire Team)

Suggested split (each member narrates):
1. **Member A (Backend):** Geo API endpoints (`/api/geocode`, `/api/hospitals/near`) and JWT/RBAC integration.
2. **Member B (Frontend):** `patient/find-care` UX (search + use my location) and Leaflet/OSM rendering.
3. **Member C (Testing):** Show results of:
   - `go test ./...`
   - `ng test --watch=false --browsers=ChromeHeadless`
   - `cypress run --headless ...`

