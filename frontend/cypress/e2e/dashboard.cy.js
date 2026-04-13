/** S4 — Karthik — role dashboards */
const API = 'http://localhost:8080';

describe('Dashboards (S4 Karthik)', () => {
  it('patient dashboard loads', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit('/patient/dashboard', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="patient-dashboard"]').should('be.visible');
  });
});
