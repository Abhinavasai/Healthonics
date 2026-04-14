/**
 * Feature 9 — Messaging (patient/doctor UI against real backend).
 * Prerequisites:
 *   - Backend: http://localhost:8080 (DATABASE_URL + JWT_SECRET; run `go run ./cmd/seed`)
 *   - Frontend: http://localhost:4200 with messaging routes (Sprint 3 feature 9 frontend)
 *
 * Run: npx cypress run --headless --spec "cypress/e2e/messaging.cy.js"
 */
const API = 'http://localhost:8080';

describe('Messaging (feature 9)', () => {
  it('loads messages page for patient after login via API', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      expect(user.role).to.eq('patient');

      cy.visit('/patient/messages', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="messages-page"]').should('be.visible');
    cy.get('[data-cy="messages-layout"]').should('be.visible');
  });

  it('doctor can open messages shell', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'doctor@healthonyx.demo',
      password: 'doctor123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      expect(user.role).to.eq('doctor');

      cy.visit('/doctor/messages', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="messages-page"]').should('be.visible');
  });
});
