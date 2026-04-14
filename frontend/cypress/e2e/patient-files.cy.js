/**
 * S4 — Abhinav — Patient files (list shell; upload optional when backend running).
 * Run backend + seed + `ng serve`, then: npx cypress run --headless --spec cypress/e2e/patient-files.cy.js
 */
const API = Cypress.env('API_URL');

describe('Patient files (S4 Abhinav)', () => {
  it('patient can open My files after login', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      expect(user.role).to.eq('patient');

      cy.visit('/patient/my-files', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="patient-files-page"]').should('be.visible');
  });
});
