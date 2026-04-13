describe('Account settings (S3 F01)', () => {
  it('shows profile from GET /api/me for patient', () => {
    cy.request('POST', '/api/login', {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      expect(user.role).to.eq('patient');

      cy.visit('/patient/settings', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="account-settings-page"]').should('be.visible');
    cy.get('[data-cy="account-settings-email"]').should('contain', 'patient@healthonyx.demo');
    cy.get('[data-cy="account-settings-role"]').should('contain', 'Patient');
  });
});
