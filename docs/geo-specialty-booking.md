# Geo, specialty, and slot-based booking

## Goals

1. **Patient** chooses a **location** (typed address / city or **browser geolocation** with consent).
2. Patient sets a **search radius** (km) and sees **hospitals** within that radius (Haversine distance on stored lat/lng).
3. Patient describes a **problem** (free text) or picks a **specialty**; the app lists **doctors** in range whose **specialization** matches, so they can open the right provider.
4. **Doctors** publish **availability slots** (start/end). **Patients** book a slot; once booked, that slot is **locked** (no double booking) and an **appointment** row is created (still subject to existing approve/reject flow).

## Data model (backend)

- **`hospitals`**: `name`, `city`, `region`, `latitude`, `longitude`.
- **`users`** (doctors): `specialization`, `hospital_id` (optional anchor), `practice_latitude`, `practice_longitude` (optional if not tied to a hospital row).
- **`doctor_slots`**: `doctor_id`, `start_at`, `end_at`, `patient_id` (NULL = open), unique `(doctor_id, start_at)`.
- **`appointments`**: optional **`slot_id`** FK to `doctor_slots` when booked via slot flow.

## API (implemented on `s2/geo-specialty-slots-backend`)

| Method | Path | Role | Purpose |
| ------ | ---- | ---- | ------- |
| GET | `/api/hospitals/near?lat=&lng=&radius_km=` | patient | Hospitals within radius |
| GET | `/api/doctors/search?lat=&lng=&radius_km=&specialization=` | patient | Doctors in range + specialty filter |
| GET | `/api/doctors/:id/slots` | patient | Open slots for a doctor |
| POST | `/api/doctor/slots` | doctor | Create availability window |
| DELETE | `/api/doctor/slots/:id` | doctor | Remove an open slot |
| POST | `/api/appointments/book-slot` | patient | Book slot → creates appointment + locks slot |

## Frontend branch (`s2/geo-specialty-slots-frontend`)

Wire **Geolocation API**, map radius UI, hospital list, specialty/problem → doctor search, slot picker, and doctor availability management. See `docs/geo-specialty-booking-frontend.md`.

## Merge notes

Merge **backend** before **frontend** so APIs exist when the UI calls them. Alternatively develop against a running backend branch locally.
