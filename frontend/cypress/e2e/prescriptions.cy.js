/**
 * S4 — Abhinav — Patient prescriptions list.
 * npx cypress run --headless --spec cypress/e2e/prescriptions.cy.js
 */
const API = 'http://localhost:8080';

describe('Prescriptions (S4 Abhinav)', () => {
  it('patient can open prescriptions page', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      cy.visit('/patient/prescriptions', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="patient-prescriptions-page"]').should('be.visible');
  });
});
