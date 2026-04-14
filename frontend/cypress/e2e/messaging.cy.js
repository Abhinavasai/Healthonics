/**
 * Feature 9 — Messaging (patient/doctor UI against real backend).
 * Prerequisites:
 *   - Backend: Cypress.env('API_URL') (default 127.0.0.1:8080; `npm run cypress:e2e` uses 18080)
 *   - Frontend: baseUrl (e.g. 4200 or 4300 for `cypress:e2e`)
 *
 * Run: npx cypress run --headless --spec "cypress/e2e/messaging.cy.js"
 */
const API = Cypress.env('API_URL');

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
