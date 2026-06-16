# Frontend Demo Script (Simple English)

Use this as a speaking script when presenting UI progress and feature flow.

---

## 1) Opening statement (30-60 seconds)

"In Sprint 2, the frontend was mainly a patient find-care experience on top of login.  
Now we have a role-based UI shell for patient, doctor, and admin, each with dedicated screens and guarded routes.  
I will walk through the UI by role and explain what has been developed from Sprint 2 to current progress."

---

## 2) Sprint 2 vs current frontend progress

### Sprint 2 UI baseline

- Login for test flow.
- Patient location search + nearby hospitals + doctor search.

### Current frontend UI scope

- Role-based app shell and protected routing.
- Patient dashboard and doctor dashboard.
- Patient appointments list + detail.
- Doctor appointments list + detail.
- Doctor availability screen.
- Patient documents and doctor document views.
- Patient files page.
- Prescriptions page.
- Messaging UI (thread + message flow).
- Notifications inbox.
- Admin shell with audit and knowledge sections.

---

## 3) Routing and shell explanation (simple)

"The frontend uses Angular routing with guards.  
After login, users are redirected to role-specific routes under one app shell.
Only authorized roles can open their pages."

### Key route groups

- Patient routes: `/patient/*`
- Doctor routes: `/doctor/*`
- Admin routes: `/admin/*`

### Guard behavior

- `authGuard`: must be logged in.
- `roleGuard`: must match route role list.

---

## 4) Patient UI walkthrough script

Say this while navigating:

1. **Dashboard**
   - "Patient sees summary cards like upcoming and pending request counts."
2. **Find Care**
   - "Patient searches by location and finds hospitals and doctors."
3. **Appointments**
   - "Patient can create requests, open detail pages, and track status."
4. **Documents**
   - "Patient can manage own document area."
5. **My Files**
   - "Patient sees uploaded files in one place."
6. **Prescriptions**
   - "Patient views active and historical prescription data."
7. **Messages**
   - "Patient can open threads and communicate with assigned doctors."
8. **Notifications**
   - "Patient sees in-app notifications in inbox style."

---

## 5) Doctor UI walkthrough script

1. **Dashboard**
   - "Doctor sees today and pending summary."
2. **Availability**
   - "Doctor manages slot availability."
3. **Appointments**
   - "Doctor opens appointment queue and detail screens."
4. **Patient Documents**
   - "Doctor views patient docs list and document detail."
   - "AI summary action is exposed in UI as assistive workflow."
5. **Messages**
   - "Doctor can continue patient threads."
6. **Notifications**
   - "Doctor sees role-specific notifications."

---

## 6) Admin UI walkthrough script

1. **Admin home**
   - "Entry screen for admin operations."
2. **Audit**
   - "Admin views audit log entries."
3. **Knowledge**
   - "Admin manages knowledge documents."

Say:

"This is a lightweight admin MVP. Advanced global settings and enterprise controls are still planned."

---

## 7) UI/UX quality points to mention

- Unified app-shell navigation for all roles.
- Guarded routes reduce unauthorized page access.
- Consistent page patterns: list -> detail -> action.
- Data-cy-friendly structure supports Cypress automation.
- Responsive behavior is included in major feature pages.

---

## 8) Frontend run and test commands

From repo root:

```powershell
cd frontend
npm install
npm start
```

Open:

- `http://localhost:4200`

Frontend tests:

```powershell
npm run test:ci
npm run build
```

Full E2E from frontend folder:

```powershell
npm run cypress:e2e
```

---

## 9) What is complete vs what is still planned (frontend)

### Strongly complete in UI

- Role shell + route protection.
- Core dashboards.
- Appointment and messaging flows.
- Find-care and slot interactions.
- Patient files and prescriptions views.
- Admin audit + knowledge basic UI.

### Partial or planned

- Rich analytics widgets and charts.
- Advanced calendar polish.
- Real-time messaging updates (websocket style).
- Full AI-driven UI modules (RAG answers, conflict review panels, AI governance controls).
- Extended enterprise admin and delivery monitoring screens.

---

## 10) Closing statement

"Frontend has evolved from a single Sprint 2 find-care flow into a full multi-role clinic portal UI foundation.  
Current work already supports major operational workflows for patients, doctors, and admins.  
Next phase focuses on deeper analytics, AI-facing interfaces, and enterprise-grade admin experiences."

