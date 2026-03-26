# Frontend: geo, specialty, and slot booking

This branch wires the UX described in `docs/geo-specialty-booking.md` to the APIs on `s2/geo-specialty-slots-backend`.

## Patient flow (target)

1. **Location** — Use the [Geolocation API](https://developer.mozilla.org/en-US/docs/Web/API/Geolocation_API) (`navigator.geolocation.getCurrentPosition`) after user consent, or manual lat/lng (address geocoding can be a follow-up).
2. **Radius** — Slider or numeric input (km); call `GET /api/hospitals/near?lat=&lng=&radius_km=`.
3. **Problem / specialty** — Text field or dropdown; map to `specialization` query on `GET /api/doctors/search?...&specialization=`.
4. **Doctors** — Show list with distance; link to **doctor detail** or straight to **slot list** via `GET /api/doctors/:id/slots`.
5. **Book** — Patient picks an open slot and submits reason; `POST /api/appointments/book-slot` with `{ slot_id, reason }`. On success, redirect to existing patient appointments.

## Doctor flow (target)

1. **Availability** — Form for `start_at` / `end_at` (ISO-8601); `POST /api/doctor/slots`.
2. **Manage** — List own slots; `DELETE /api/doctor/slots/:id` for open slots.

## This branch (scaffolding)

- `GeoBookingService` wraps the HTTP calls.
- Placeholder routes: `/patient/find-care`, `/doctor/availability` (shell nav when logged in).
- Full UI (maps, geolocation prompts, validation) can be layered in follow-up commits.

**Merge order:** land backend branch first, then frontend, or test locally with backend running from its branch.
