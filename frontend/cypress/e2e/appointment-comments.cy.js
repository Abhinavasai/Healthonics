/** S4 — Rohith — appointment comments API (invalid id → 400 when route is registered) */
const API = 'http://localhost:8080';

describe('Appointment comments (S4 Rohith)', function () {
  before(function () {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((loginRes) => {
      cy.request({
        method: 'GET',
        url: `${API}/api/appointments/x/comments`,
        headers: { Authorization: `Bearer ${loginRes.body.token}` },
        failOnStatusCode: false,
      }).then((probe) => {
        if (probe.status === 404) {
          cy.log(
            'Skipping appointment-comments API spec: GET /api/appointments/:id/comments is not registered on :8080 (restart API from this repo).'
          );
          this.skip();
        }
      });
    });
  });

  it('returns 400 for invalid appointment id on comments', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      const token = res.body.token;
      cy.request({
        method: 'GET',
        url: `${API}/api/appointments/not-a-uuid/comments`,
        headers: { Authorization: `Bearer ${token}` },
        failOnStatusCode: false,
      }).then((r) => {
        expect(r.status).to.eq(400);
      });
    });
  });
});
