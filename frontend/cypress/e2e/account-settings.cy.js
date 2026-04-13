describe('Account settings (S3 F01)', () => {
  it('shows profile from GET /api/me for patient', () => {
    cy.intercept('GET', '**/api/me', {
      id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
      email: 'pat@settings.demo',
      role: 'patient',
    }).as('me');

    cy.visit('/patient/settings', {
      onBeforeLoad(win) {
        win.localStorage.setItem('token', 'cypress-token');
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
            email: 'pat@settings.demo',
            role: 'patient',
          })
        );
      },
    });

    cy.wait('@me');
    cy.get('[data-cy="account-settings-page"]').should('be.visible');
    cy.get('[data-cy="account-settings-email"]').should('contain', 'pat@settings.demo');
    cy.get('[data-cy="account-settings-role"]').should('contain', 'Patient');
  });
});
