/** S4 — Rohith — appointment comments API */
const API = 'http://localhost:8080';

describe('Appointment comments (S4 Rohith)', () => {
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
