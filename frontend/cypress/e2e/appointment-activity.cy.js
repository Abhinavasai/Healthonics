describe('Appointment detail + activity timeline', () => {
  it('doctor detail shows activity rows from API', () => {
    cy.intercept('GET', '**/api/appointments/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', {
      id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
      patient_id: '22222222-2222-2222-2222-222222222222',
      doctor_id: '33333333-3333-3333-3333-333333333333',
      scheduled_at: '2026-06-01T14:00:00Z',
      reason: 'Follow-up',
      status: 'approved',
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    }).as('appt');

    cy.intercept('GET', '**/api/appointments/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/activity', {
      activities: [
        {
          id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
          appointment_id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
          actor_user_id: '33333333-3333-3333-3333-333333333333',
          actor_email: 'doctor@healthonyx.demo',
          action: 'status_changed',
          detail: 'approved',
          created_at: '2026-05-02T10:00:00Z',
        },
      ],
    }).as('act');

    cy.visit('/doctor/appointments/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', {
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

    cy.wait('@appt');
    cy.wait('@act');
    cy.get('[data-cy="doctor-appointment-detail-page"]').should('be.visible');
    cy.contains('Follow-up').should('be.visible');
    cy.get('[data-cy="appointment-activity-row"]').should('have.length', 1);
    cy.contains('Request approved').should('be.visible');
    cy.get('[data-cy="appointment-activity-actor"]').should('contain', 'doctor@healthonyx.demo');
    // Top-layer backdrop (e.g. browser popover layer) can intercept naive clicks in headless runs.
    cy.get('[data-cy="appointment-activity-details"]').scrollIntoView().click({ force: true });
    cy.get('[data-cy="appointment-activity-expanded"]').should('be.visible');
    cy.get('[data-cy="appointment-activity-refresh"]').should('be.visible').click();
    cy.wait('@act');
  });

  it('patient detail shows empty activity message when none', () => {
    cy.intercept('GET', '**/api/appointments/cccccccc-cccc-cccc-cccc-cccccccccccc', {
      id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
      patient_id: '22222222-2222-2222-2222-222222222222',
      doctor_id: '33333333-3333-3333-3333-333333333333',
      scheduled_at: '2026-06-01T14:00:00Z',
      reason: 'Checkup',
      status: 'pending',
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-01T10:00:00Z',
    }).as('appt2');

    cy.intercept('GET', '**/api/appointments/cccccccc-cccc-cccc-cccc-cccccccccccc/activity', {
      activities: [],
    }).as('act2');

    cy.visit('/patient/appointments/cccccccc-cccc-cccc-cccc-cccccccccccc', {
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

    cy.wait('@appt2');
    cy.wait('@act2');
    cy.get('[data-cy="patient-appointment-detail-page"]').should('be.visible');
    cy.get('[data-cy="appointment-activity-empty"]').should('be.visible');
  });
});
