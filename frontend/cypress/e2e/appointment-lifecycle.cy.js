/** S4 — Kaushik — appointment lifecycle UI shell */
const API = Cypress.env('API_URL');

describe('Appointment lifecycle (S4 Kaushik)', () => {
  it('patient appointments list is reachable', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit('/patient/appointments', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('.appointments h1').should('contain', 'Patient Appointments');
  });
});
