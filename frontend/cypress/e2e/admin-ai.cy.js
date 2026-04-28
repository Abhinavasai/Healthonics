describe('admin ai observability', () => {
  it('renders controls and runs eval', () => {
    const API = Cypress.env('API_URL');
    cy.intercept('GET', '**/api/admin/ai/observability', {
      statusCode: 200,
      body: {
        settings: {
          ollama_enabled: false,
          fallback_enabled: true,
          rate_limit_enabled: true,
          rate_limit_per_minute: 30,
          cache_enabled: true,
          cache_ttl_seconds: 900,
          ollama_model: 'llama3.1:8b',
          updated_at: '2026-04-28T00:00:00Z',
          updated_by: ''
        },
        observability: {
          queued_jobs_24h: 4,
          completed_jobs_24h: 3,
          failed_jobs_24h: 1,
          avg_latency_seconds: 0.35,
          cache_ready_count: 10,
          pending_docs_count: 2,
          failed_docs_count: 1
        }
      }
    }).as('getObs');

    cy.intercept('PUT', '**/api/admin/ai/settings', (req) => {
      req.reply({
        statusCode: 200,
        body: {
          settings: {
            ...req.body,
            updated_at: '2026-04-28T00:01:00Z',
            updated_by: 'admin-id'
          },
          observability: {
            queued_jobs_24h: 5,
            completed_jobs_24h: 4,
            failed_jobs_24h: 1,
            avg_latency_seconds: 0.31,
            cache_ready_count: 11,
            pending_docs_count: 1,
            failed_docs_count: 1
          }
        }
      });
    }).as('putSettings');

    cy.intercept('POST', '**/api/admin/ai/eval', {
      statusCode: 200,
      body: {
        eval: {
          model_name: 'llama3.1:8b',
          samples_evaluated: 20,
          success_rate: 0.8,
          quality_score: 80
        }
      }
    }).as('runEval');

    cy.request('POST', `${API}/api/login`, {
      email: 'admin@healthonyx.demo',
      password: 'admin123'
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit('/admin/ai', {
        onBeforeLoad(win) {
          win.localStorage.setItem('token', token);
          win.localStorage.setItem('user', JSON.stringify(user));
        }
      });
    });
    cy.wait('@getObs');

    cy.contains('AI observability and controls').should('exist');
    cy.get('input[name="rateLimitPerMinute"]').clear().type('25');
    cy.contains('Save controls').click();
    cy.wait('@putSettings');
    cy.contains('AI runtime settings updated.').should('exist');

    cy.contains('Run eval').click();
    cy.wait('@runEval');
    cy.contains('Eval result').should('exist');
    cy.contains('Samples:').should('exist');
  });
});
