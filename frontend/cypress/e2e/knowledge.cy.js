/** S4 — Rohith — admin knowledge base */
const API = 'http://localhost:8080';

describe('Knowledge base (S4 Rohith)', () => {
  it('admin can open knowledge page', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'admin@healthonyx.demo',
      password: 'admin123',
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit('/admin/knowledge', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="admin-knowledge-page"]').should('be.visible');
  });
});
