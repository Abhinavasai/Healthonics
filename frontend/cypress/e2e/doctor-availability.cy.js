/** Doctor availability: calendar + time-only UI */
const API = Cypress.env('API_URL');

function login(email, password) {
  return cy.request('POST', `${API}/api/login`, { email, password }).then((res) => res.body);
}

describe('Doctor availability UI (calendar + time inputs)', () => {
  const randomStartTime = () => {
    // 08:00 - 17:30 in 30-minute increments, pseudo-random per run.
    const candidates = [];
    for (let h = 8; h <= 17; h += 1) {
      candidates.push(`${String(h).padStart(2, '0')}:00`);
      if (h !== 17) {
        candidates.push(`${String(h).padStart(2, '0')}:30`);
      }
    }
    const idx = Date.now() % candidates.length;
    return candidates[idx];
  };

  it('doctor can create a slot using date+time UI (UTC)', () => {
    login('doctor@healthonyx.demo', 'doctor123').then(({ token, user }) => {
      cy.visit('/doctor/availability', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    const tomorrow = new Date();
    tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
    const dateStr = tomorrow.toISOString().slice(0, 10); // YYYY-MM-DD

    const startTime = randomStartTime();
    cy.get('[data-cy="doctor-availability-date"]').clear().type(dateStr);
    cy.get('[data-cy="doctor-availability-start-time"]').clear().type(startTime);

    cy.get('[data-cy="doctor-availability-add-slot"]').click();
    cy.contains(/Slot created|already exists/, { timeout: 10000 }).should('exist');
    cy.get('tr[data-cy="doctor-slot-row"]', { timeout: 10000 }).should('have.length.at.least', 1);
  });

  it('doctor can delete an open slot from the table', () => {
    login('doctor@healthonyx.demo', 'doctor123').then(({ token, user }) => {
      cy.visit('/doctor/availability', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        },
      });
    });

    // Create a known slot first so delete assertion is deterministic.
    const tomorrow = new Date();
    tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
    const dateStr = tomorrow.toISOString().slice(0, 10);
    const startTime = '17:30';
    cy.get('[data-cy="doctor-availability-date"]').clear().type(dateStr);
    cy.get('[data-cy="doctor-availability-start-time"]').clear().type(startTime);
    cy.get('[data-cy="doctor-availability-add-slot"]').click();
    cy.contains(/Slot created|already exists/, { timeout: 10000 }).should('exist');

    let startAtBefore = '';
    cy.get('tr[data-cy="doctor-slot-row"]', { timeout: 10000 }).then(($rows) => {
      const targetRow = [...$rows].find((r) => (r.getAttribute('data-start-at') || '').includes('T17:30:'));
      if (targetRow) {
        startAtBefore = targetRow.getAttribute('data-start-at') || '';
        cy.wrap(targetRow).find('[data-cy="doctor-slot-delete"]').click();
        return;
      }

      const availableRow = [...$rows].find((r) => r.getAttribute('data-available') === 'true');
      expect(availableRow, 'expected at least one available slot').to.exist;
      startAtBefore = availableRow?.getAttribute('data-start-at') || '';
      cy.wrap(availableRow).find('[data-cy="doctor-slot-delete"]').click();
    });

    // Eventually that start_at should no longer exist in rows.
    cy.get('tr[data-cy="doctor-slot-row"]', { timeout: 10000 }).should(($rows) => {
      const stillThere = [...$rows].some((r) => (r.getAttribute('data-start-at') || '') === startAtBefore);
      expect(stillThere).to.equal(false);
    });
  });
});

