/**
 * Sprint 3.5 — API smoke via dev-server proxy (same origin as Cypress baseUrl).
 * Run with: backend on :8080, `ng serve` on :4200.
 */
describe('Sprint 3.5 API smoke (proxied)', () => {
  it('GET /api/dashboard/patient returns counts', () => {
    cy.request('POST', '/api/login', {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((login) => {
      expect(login.status).to.eq(200);
      const { token } = login.body;
      cy.request({
        method: 'GET',
        url: '/api/dashboard/patient',
        headers: { Authorization: `Bearer ${token}` },
      }).then((res) => {
        expect(res.status).to.eq(200);
        expect(res.body).to.have.property('upcoming_count');
        expect(res.body).to.have.property('pending_count');
        expect(res.body).to.have.property('unread_messages');
      });
    });
  });

  it('GET /api/me/notification-preferences', () => {
    cy.request('POST', '/api/login', {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((login) => {
      const { token } = login.body;
      cy.request({
        method: 'GET',
        url: '/api/me/notification-preferences',
        headers: { Authorization: `Bearer ${token}` },
      }).then((res) => {
        expect(res.status).to.eq(200);
        expect(res.body).to.have.property('email_appointment_reminders');
        expect(res.body).to.have.property('sms_appointment_reminders');
      });
    });
  });
});
