/** S4 — Karthik — role dashboards */
const API = Cypress.env('API_URL');

describe('Dashboards (S4 Karthik)', () => {
  it('patient dashboard loads', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'patient@healthonyx.demo',
      password: 'patient123',
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit('/patient/dashboard', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });
    cy.get('[data-cy="patient-dashboard"]').should('be.visible');
  });

  it('doctor can acknowledge a critical escalation', () => {
    cy.request('POST', `${API}/api/login`, {
      email: 'doctor@healthonyx.demo',
      password: 'doctor123',
    }).then((res) => {
      const { token, user } = res.body;
      cy.intercept('GET', '/api/doctor/dashboard/summary', {
        statusCode: 200,
        body: {
          role: 'doctor',
          appointments_today: 3,
          pending_queue: 1,
          aggregations: { unread_messages: 2, appointments_this_week: 9, critical_open_count: 1 },
          alerts: [{ severity: 'warning', code: 'critical_result_escalations', message: 'Critical findings require acknowledgement.', count: 1 }],
          critical_escalations: [
            {
              id: 'esc-1',
              document_id: 'doc-1',
              patient_id: 'pat-1',
              patient_email: 'patient@healthonyx.demo',
              filename: 'lab-result.pdf',
              tier: 2,
              status: 'open',
              rule_code: 'critical_summary_keyword',
              message: 'Tier 2 critical result: lab-result.pdf for patient patient@healthonyx.demo requires review.',
              first_detected_at: '2026-04-28T20:00:00Z',
              next_escalation_at: '2026-04-28T20:15:00Z',
            },
          ],
        },
      }).as('doctorSummary');

      cy.intercept('POST', '/api/doctor/critical-escalations/esc-1/ack', {
        statusCode: 200,
        body: { id: 'esc-1', status: 'acknowledged' },
      }).as('ackEsc');

      cy.visit('/doctor/dashboard', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    cy.wait('@doctorSummary');
    cy.get('[data-cy="doctor-critical-escalations"]').should('be.visible');
    cy.get('[data-cy="ack-critical-esc-1"]').click();
    cy.wait('@ackEsc');
    cy.get('[data-cy="ack-critical-esc-1"]').should('be.disabled');
  });
});
