/** AI summary retry path: validate retry CTA stability */
const API = Cypress.env('API_URL');

function login(email, password) {
  return cy.request('POST', `${API}/api/login`, { email, password }).then((res) => res.body);
}

describe('AI summary retry path', () => {
  it('doctor can trigger summary and recover from transient request error', () => {
    login('doctor@healthonyx.demo', 'doctor123').then(({ token, user }) => {
      cy.visit('/doctor/documents', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="doctor-documents-open"]').first().click();

    // Simulate one transient backend failure then success on retry.
    let callCount = 0;
    cy.intercept('POST', '**/api/doctor/documents/*/summarize', (req) => {
      callCount += 1;
      if (callCount === 1) {
        req.reply({
          statusCode: 500,
          body: { error: 'Temporary summarization outage' },
        });
        return;
      }
      req.reply({
        statusCode: 202,
        body: { status: 'pending', message: 'summary job already queued' },
      });
    }).as('summarizeReq');

    cy.get('[data-cy="doctor-document-summarize"]').click();
    cy.wait('@summarizeReq');

    cy.get('[data-cy="doctor-document-summary-request-error"]').should('contain.text', 'Could not start summarization');
    cy.get('[data-cy="doctor-document-summary-retry-from-error"]').click();
    cy.wait('@summarizeReq');

    cy.get('[data-cy="doctor-document-summary-request-error"]').should('not.exist');
    cy.get('[data-cy="doctor-document-summary-info"]').should('exist');
  });
});

