describe('Patient find-care (location -> nearby hospitals)', () => {
  it('loads and shows hospitals after geocoding + search', () => {
    // Stub auth (guards use localStorage token + user.role)
    cy.visit('/patient/find-care', {
      onBeforeLoad(win) {
        win.localStorage.setItem('token', 'cypress-token');
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            id: 'cypress-user-id',
            email: 'patient@healthonyx.demo',
            role: 'patient',
          })
        );
      },
    });

    cy.intercept('GET', '**/api/geocode*', {
      results: [
        {
          lat: 29.6516,
          lng: -82.3248,
          display_name: 'Gainesville, FL',
        },
      ],
    }).as('geocode');

    cy.intercept('GET', '**/api/hospitals/near*', {
      hospitals: [
        {
          id: '11111111-1111-1111-1111-111111111111',
          name: 'UF Health Shands Hospital',
          city: 'Gainesville',
          region: 'Florida',
          latitude: 29.6516,
          longitude: -82.3248,
          distance_km: 1.23,
        },
      ],
    }).as('hospitalsNear');

    // Search for a location (triggers /api/geocode) then show nearby hospitals
    cy.get('input[type="text"]').first().clear().type('Gainesville, FL');
    cy.contains('button', 'Search').click();
    cy.wait('@geocode');

    cy.contains('button', 'Hospitals nearby').click();
    cy.wait('@hospitalsNear');

    cy.contains('UF Health Shands Hospital').should('be.visible');
  });
});

