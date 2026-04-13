/**
 * Sprint 3.5 — notification preferences on account settings (patient).
 */
describe('Sprint 3.5 notification preferences', () => {
  it('shows notification preferences for patient', () => {
    cy.request('POST', '/api/login', {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      cy.visit('/patient/settings', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="notification-preferences-card"]').should('be.visible');
    cy.get('[data-cy="pref-email-reminders"]').should('exist');
    cy.get('[data-cy="pref-sms-reminders"]').should('exist');
  });
});
