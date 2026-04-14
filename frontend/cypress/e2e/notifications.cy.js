/** S4 — Kaushik — notifications inbox */
const API = Cypress.env('API_URL');

describe('Notifications inbox (S4 Kaushik)', () => {
  it('patient opens notifications', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit('/patient/notifications', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="notifications-inbox-page"]').should('be.visible');
  });
});
