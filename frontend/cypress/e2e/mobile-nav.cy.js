/**
 * Sprint 3 F02 — mobile shell navigation (drawer + backdrop).
 */
describe('Mobile app shell navigation', () => {
  it('opens and closes sidebar via menu, backdrop, and nav link', () => {
    cy.intercept('GET', '**/api/appointments/patient', { appointments: [] }).as('appts');
    cy.intercept('GET', '**/api/doctors', { doctors: [] }).as('docs');

    cy.viewport(390, 844);

    cy.request('POST', '/api/login', {
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

    cy.url().should('include', '/patient/appointments');
    cy.get('app-shell').should('exist');
    cy.get('[data-cy="mobile-nav-toggle"]').should('be.visible').click();
    cy.get('#app-sidebar').should('have.class', 'open');
    cy.get('[data-cy="mobile-nav-backdrop"]').should('have.class', 'nav-backdrop--visible');

    cy.get('[data-cy="mobile-nav-backdrop"]').click({ force: true });
    cy.get('#app-sidebar').should('not.have.class', 'open');

    cy.get('[data-cy="mobile-nav-toggle"]').click();
    cy.contains('a.nav-item', 'Find care').click();
    cy.get('#app-sidebar').should('not.have.class', 'open');
  });
});
