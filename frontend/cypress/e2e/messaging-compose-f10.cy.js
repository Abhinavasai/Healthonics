/**
 * Sprint 3 feature 10 — messaging compose UX (limits UI), no real backend.
 */
describe('Messaging compose limits (F10)', () => {
  const threadId = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';

  it('disables reply send when text exceeds 8000 runes', () => {
    cy.intercept('GET', '**/api/doctors', { doctors: [] }).as('doctors');
    cy.intercept('GET', '**/api/messages/unread', { unread_total: 0 }).as('unread');
    cy.intercept('GET', '**/api/messages/threads', {
      threads: [
        {
          id: threadId,
          peer_user_id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
          peer_email: 'doc@demo.test',
          last_message_at: '2026-05-01T10:00:00Z',
          last_preview: 'Hi',
          unread_count: 0,
        },
      ],
    }).as('threads');
    cy.intercept('GET', `**/api/messages/threads/${threadId}`, {
      messages: [],
    }).as('msgs');
    cy.intercept('POST', `**/api/messages/threads/${threadId}/read`, {}).as('read');

    cy.visit(`/patient/messages/${threadId}`, {
      onBeforeLoad(win) {
        win.localStorage.setItem('token', 'cypress-token');
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
            email: 'pat@demo.test',
            role: 'patient',
          })
        );
      },
    });

    cy.wait('@threads');
    cy.get('[data-cy="messages-reply-input"]').should('be.visible');
    const over = 'm'.repeat(8001);
    cy.get('[data-cy="messages-reply-input"]').invoke('val', over).trigger('input', { force: true });
    cy.get('[data-cy="messages-reply-count"]').should('contain', '8001');
    cy.get('[data-cy="messages-reply-count"]').should('have.class', 'char-count--bad');
    cy.get('[data-cy="messages-send-reply"]').should('be.disabled');
  });
});
