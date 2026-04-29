/** AI usability: doctor document summary flow (fallback/local extraction) */
const API = Cypress.env('API_URL');

function login(email, password) {
  return cy.request('POST', `${API}/api/login`, { email, password }).then((res) => res.body);
}

describe('AI document summary usability', () => {
  it('doctor can run AI summary and see updated summary text', () => {
    // Ensure the doctor summary endpoint doesn't short-circuit via cache.
    login('admin@healthonyx.demo', 'admin123')
      .then(({ token }) => {
        return cy.request({
          method: 'PUT',
          url: `${API}/api/admin/ai/settings`,
          headers: {
            Authorization: `Bearer ${token}`,
          },
          body: {
            ollama_enabled: false,
            fallback_enabled: true,
            rate_limit_enabled: false,
            rate_limit_per_minute: 10000,
            cache_enabled: false,
            cache_ttl_seconds: 900,
            ollama_model: 'llama3.1:8b',
            prompt_version: 'v1',
          },
        });
      })
      .then(() => {
        return login('doctor@healthonyx.demo', 'doctor123').then(({ token, user }) => {
          cy.visit('/doctor/documents', {
            onBeforeLoad(win) {
              win.localStorage.setItem('token', token);
              win.localStorage.setItem('user', JSON.stringify(user));
            },
          });
        });
      });

    cy.get('[data-cy="doctor-documents-row"]').should('have.length.at.least', 1);
    cy.get('[data-cy="doctor-documents-open"]').first().click();

    cy.get('[data-cy="doctor-document-summarize"]').should('exist');
    cy.get('[data-cy="doctor-document-summarize"]').click();

    cy.get('[data-cy="doctor-document-summary-pending"]', { timeout: 20000 }).should('exist');

    cy.get('[data-cy="doctor-document-summary-text"]', { timeout: 60000 }).should(($el) => {
      const txt = $el.text();
      expect(txt).to.contain('AI summary (local extraction');
    });
  });
});

