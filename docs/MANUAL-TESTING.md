# Healthonyx — Manual testing guide

Hands-on commands for running the stack locally, then feature-by-feature steps and expected outcomes.

---

## 1. One-time prerequisites

- **PostgreSQL** running and a **database URL** you can use.
- **Go** and **Node.js** installed.
- Create `backend/.env` (or set environment variables) with at least:

```env
DATABASE_URL=postgres://USER:PASSWORD@localhost:5432/healthonyx?sslmode=disable
JWT_SECRET=your-long-random-secret
PORT=8080
```

Optional for geocoding (Nominatim):

```env
GEOCODE_USER_AGENT=HealthonyxLocalDev/1.0 (your@email)
```

---

## 2. Commands to run (typical local session)

### Terminal 1 — database + migrate + seed (once per fresh DB)

```powershell
cd path\to\Healthonyx\backend\cmd\seed
go run .
```

You should see output like seeded `patient@healthonyx.demo`, `doctor@healthonyx.demo`, `admin@healthonyx.demo` with passwords printed.

### Terminal 2 — API

```powershell
cd path\to\Healthonyx\backend
go run .
```

Expect a log line such as: **Listening on :8080** (or your `PORT`).

### Terminal 3 — Angular (proxies `/api` to backend)

```powershell
cd path\to\Healthonyx\frontend
npm install
ng serve --host 127.0.0.1 --port 4200
```

Open: **http://localhost:4200**

### Optional — API smoke (same machine)

From the repo root:

```powershell
npx newman run docs\Healthonyx-API-Sprint3-Implemented.postman_collection.json
npx newman run docs\Healthonyx-API-Complete-Regression.postman_collection.json
```

### Optional — E2E (with API + `ng serve` running)

```powershell
cd path\to\Healthonyx\frontend
npm run cypress:open
```

Or headless (if your project defines it):

```powershell
npm run cypress:run
```

---

## 3. Demo accounts (from seed)

| Role    | Email                     | Password     |
|---------|---------------------------|--------------|
| Patient | `patient@healthonyx.demo` | `patient123` |
| Doctor  | `doctor@healthonyx.demo`  | `doctor123`  |
| Admin   | `admin@healthonyx.demo`   | `admin123`   |

---

## 4. Feature-by-feature manual testing

### A) Health and API reachability

1. Browser or curl: `GET http://localhost:8080/health`  
   **Expect:** `200` JSON such as `{ "status": "ok" }`.

2. With the app open, DevTools → Network: after login, `/api/...` calls should go to the **same origin** (`4200`) and proxy to **8080**.

---

### B) Registration and login

1. **Register** (`/register`): create a new patient (unique email).  
   **Expect:** success and redirect or onboarding flow as implemented.

2. **Login** (`/login`) as each demo user.  
   **Expect:**
   - Patient → `/patient` → **dashboard** (`/patient/dashboard`)
   - Doctor → `/doctor/dashboard`
   - Admin → `/admin`

3. Wrong password.  
   **Expect:** error message; remain on login.

---

### C) Account / profile (`/patient/settings`, `/doctor/settings`)

1. Open **Account**.  
   **Expect:** email, role, user id; **Refresh** reloads from the server.

2. **Notification preferences** (patient/doctor; admin may differ): toggles and **Save**.  
   **Expect:** save succeeds; reload still shows saved values (persisted in DB).

---

### D) Patient dashboard (`/patient/dashboard`)

**Expect (typical with seed data):**

- Cards: **Upcoming**, **Pending approval**, **Unread messages** (numbers ≥ 0).
- **Next visit** block if a future appointment exists.
- Links to Find care, Appointments, Messages work.

---

### E) Doctor dashboard (`/doctor/dashboard`)

**Expect:**

- Today’s appointments, pending queue, unread messages — numeric summaries without hard errors.
- Nav: Availability, Appointments, Messages, Settings.

---

### F) Admin dashboard (`/admin`)

**Expect:**

- **Database:** OK (when DB is up).
- Counts for patients, doctors, admins.
- **Users** table: email, role, active; **Disable/Enable** updates state (avoid disabling your only admin in real environments).

---

### G) Find care (`/patient/find-care`)

1. Enter a location (for example `Gainesville, FL`) → **Search**.  
   **Expect:** suggestions list.

2. **Hospitals nearby**.  
   **Expect:** list of hospitals with distances.

3. Pick a **hospital** (optional) → **Department** list reflects that hospital.  
   **Expect:** departments populate or the “any” path works.

4. **Find doctors**.  
   **Expect:** doctors within radius; with hospital filter, results narrow.

5. Map loads (Leaflet).  
   **Expect:** tiles load; markers are plausible.

**Failure modes:** backend down / proxy misconfigured → empty lists or errors in the Network tab.

---

### H) Patient appointments (`/patient/appointments`)

1. **Request appointment:** choose doctor, date/time, reason → **Submit**.  
   **Expect:** new row in the list; status often **Pending**.

2. **Appointment detail** (`/patient/appointments/:id`).  
   **Expect:** scheduled time, reason, status badge, **activity timeline** (may be empty at first).

3. **Cancel** (when allowed: future slot and valid status).  
   **Expect:** status → **Cancelled**; new activity line.

4. **Request reschedule** (when allowed): future datetime + optional note.  
   **Expect:** status → **Reschedule requested**, or a clear error if rules fail (for example already cancelled).

---

### I) Doctor appointments (`/doctor/appointments` and detail)

1. List shows assigned appointments.  
   **Expect:** pending items may expose **Approve/Reject** on the list where implemented.

2. Open **detail** for **Pending**: **Approve** / **Reject**.  
   **Expect:** status updates; activity records the action.

3. For **Reschedule requested**: **Approve new time** / **Keep original**.  
   **Expect:** appointment returns to **Approved** with updated or original time per backend rules.

4. For **Approved**: **Mark completed** / **No-show** / **Cancel appointment** where applicable.  
   **Expect:** terminal statuses; activity updates.

---

### J) Doctor availability (`/doctor/availability`)

1. Create a slot; list shows open slots.  
   **Expect:** slot appears; delete removes it.

2. Booking from find-care should consume a slot when that flow is used.  
   **Expect:** slot no longer available for others.

---

### K) Messaging (`/patient/messages`, `/doctor/messages`)

1. Patient: start or open a thread with a doctor.  
   **Expect:** messages send; thread list updates; unread counts change.

2. Doctor replies; patient sees messages (refresh or polling).  
   **Expect:** two-way conversation.

3. Very long message (over any documented limit).  
   **Expect:** validation message in UI.

---

### L) Postman collections (API regression)

1. Import **`docs/Healthonyx-API-Sprint3-Implemented.postman_collection.json`** (and/or Complete Regression) into Postman.
2. Set collection variable **`baseUrl`** = `http://localhost:8080`.
3. Run **Auth** first (tokens), then the rest of the folders.

**Expect:**

- Most requests return **2xx**.
- Requests that intentionally assert errors return **400** (or whatever the test expects), for example invalid hospital UUID.

---

### M) Cypress

With backend on **8080** and **`ng serve`** on **4200**, run Cypress from `frontend`.

**Expect:** all specs pass (align with `npm run cypress:run` in CI if configured).

---

## 5. Quick troubleshooting

| Symptom | Likely cause |
|--------|----------------|
| `/api` 404 from the browser | Frontend not proxying — run `ng serve` from `frontend` with the project’s `proxy.conf.json`. |
| Login **401** / **500** | Wrong `JWT_SECRET` or `DATABASE_URL`; DB down; migrations not applied. |
| Geocode empty | Nominatim / network; set `GEOCODE_USER_AGENT`. |
| No doctors in find-care | Radius too small or seed data / hospital filter mismatch. |

---

## Related files

- `docs/Healthonyx-API-Sprint3-Implemented.postman_collection.json`
- `docs/Healthonyx-API-Complete-Regression.postman_collection.json`

Replace `path\to\Healthonyx` with your actual clone path (for example `c:\Users\kaush\Desktop\Healthonyx`).
