/** Demo data smoke checks — seeded documents/reminders/slots */
const API = Cypress.env('API_URL');

function login(email, password) {
  return cy.request('POST', `${API}/api/login`, { email, password }).then((res) => res.body);
}

describe('Demo data (documents, reminders, slots)', () => {
  it('doctor sees seeded patient documents', () => {
    login('doctor@healthonyx.demo', 'doctor123').then(({ token, user }) => {
      cy.visit('/doctor/documents', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="doctor-documents-list-page"]').should('be.visible');
    cy.get('[data-cy="doctor-documents-row"]').should('have.length.at.least', 2);
  });

  it('patient notifications inbox shows seeded items', () => {
    login('patient@healthonyx.demo', 'patient123').then(({ token, user }) => {
      cy.visit('/patient/notifications', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="notifications-inbox-page"]').should('be.visible');
    cy.get('[data-cy="notifications-list"]').should('exist');
    cy.get('[data-cy="notifications-row"]').should('have.length.at.least', 1);
  });

  it('admin knowledge base shows seeded documents', () => {
    login('admin@healthonyx.demo', 'admin123').then(({ token, user }) => {
      cy.visit('/admin/knowledge', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="admin-knowledge-page"]').should('be.visible');
    cy.get('[data-cy="knowledge-doc-row"]').should('have.length.at.least', 3);
  });

  it('doctor availability page shows seeded slots', () => {
    login('doctor@healthonyx.demo', 'doctor123').then(({ token, user }) => {
      cy.visit('/doctor/availability', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="doctor-availability-slots"]').should('be.visible');
    cy.get('[data-cy="doctor-slot-row"]').should('have.length.at.least', 1);
  });
});

