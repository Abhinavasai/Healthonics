/** S4 — Karthik — admin audit log */
const API = Cypress.env('API_URL');

describe('Admin audit (S4 Karthik)', () => {
  it('admin can open audit page', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'admin@healthonyx.demo',
      password: 'admin123',
    }).then((res) => {
      const { token, user } = res.body;
      expect(user.role).to.eq('admin');
      cy.visit('/admin/audit', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="admin-audit-page"]').should('be.visible');
  });
});
