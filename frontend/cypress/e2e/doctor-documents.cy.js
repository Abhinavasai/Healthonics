describe('Doctor patient documents (list + detail)', () => {
  const stubUser = () => {
    cy.visit('/doctor/documents', {
      onBeforeLoad(win) {
        win.localStorage.setItem('token', 'cypress-token');
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            id: '33333333-3333-3333-3333-333333333333',
            email: 'doctor@healthonyx.demo',
            role: 'doctor',
          })
        );
      },
    });
  };

  it('shows document rows when API returns data', () => {
    cy.intercept('GET', '**/api/doctor/documents', {
      documents: [
        {
          id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
          filename: 'labs.pdf',
          patient_email: 'patient@healthonyx.demo',
          created_at: '2026-04-01T12:00:00Z',
          summary_status: 'none',
        },
      ],
    }).as('docList');

    stubUser();
    cy.wait('@docList');
    cy.get('[data-cy="doctor-documents-list-page"]').should('be.visible');
    cy.get('[data-cy="doctor-documents-row"]').should('have.length', 1);
    cy.contains('labs.pdf').should('be.visible');
    cy.get('[data-cy="doctor-documents-open"]').click();
    cy.url().should('include', '/doctor/documents/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');
  });

  it('detail page shows generate summary for none status (stubbed)', () => {
    cy.intercept('GET', '**/api/doctor/documents/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', {
      id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
      filename: 'labs.pdf',
      patient_id: '22222222-2222-2222-2222-222222222222',
      patient_email: 'patient@healthonyx.demo',
      size_bytes: 1200,
      content_type: 'application/pdf',
      created_at: '2026-04-01T12:00:00Z',
      summary: null,
      summary_status: 'none',
    }).as('detail');

    cy.visit('/doctor/documents/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', {
      onBeforeLoad(win) {
        win.localStorage.setItem('token', 'cypress-token');
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            id: '33333333-3333-3333-3333-333333333333',
            email: 'doctor@healthonyx.demo',
            role: 'doctor',
          })
        );
      },
    });

    cy.wait('@detail');
    cy.contains('Generate summary').should('be.visible');
  });
});

describe('Patient document upload page', () => {
  it('submits multipart upload (stubbed)', () => {
    cy.intercept('POST', '**/api/patient/documents', {
      id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
      filename: 'note.txt',
    }).as('upload');

    cy.visit('/patient/documents', {
      onBeforeLoad(win) {
        win.localStorage.setItem('token', 'cypress-token');
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            id: '22222222-2222-2222-2222-222222222222',
            email: 'patient@healthonyx.demo',
            role: 'patient',
          })
        );
      },
    });

    cy.get('[data-cy="patient-documents-upload-page"]').should('be.visible');
    cy.get('[data-cy="patient-document-file"]').selectFile(
      {
        contents: Cypress.Buffer.from('hello'),
        fileName: 'note.txt',
        mimeType: 'text/plain',
      },
      { force: true }
    );
    cy.get('[data-cy="patient-document-upload"]').click();
    cy.wait('@upload');
    cy.get('[data-cy="patient-document-success"]').should('be.visible');
  });
});
