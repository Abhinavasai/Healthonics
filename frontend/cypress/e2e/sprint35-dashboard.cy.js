/**
 * Sprint 3.5 — role dashboards load with seeded demo users (requires backend + `ng serve` with proxy).
 */
describe('Sprint 3.5 dashboards', () => {
  it('patient dashboard shows summary cards', () => {
    cy.request('POST', '/api/login', {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      cy.visit('/patient/dashboard', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('app-patient-dashboard .dash h1').should('contain', 'Dashboard');
    cy.get('app-patient-dashboard .grid .card').should('have.length.at.least', 3);
  });

  it('doctor dashboard loads', () => {
    cy.request('POST', '/api/login', {
      email: 'doctor@healthonyx.demo',
      password: 'doctor123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      cy.visit('/doctor/dashboard', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('app-doctor-dashboard').should('exist');
    cy.contains('Dashboard').should('be.visible');
  });

  it('admin dashboard lists users', () => {
    cy.request('POST', '/api/login', {
      email: 'admin@healthonyx.demo',
      password: 'admin123',
    }).then((res) => {
      expect(res.status).to.eq(200);
      const { token, user } = res.body;
      cy.visit('/admin', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('app-admin-dashboard').should('exist');
    cy.contains('Users').should('be.visible');
  });
});
