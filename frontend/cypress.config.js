const { defineConfig } = require('cypress');

module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://localhost:4200',
    env: {
      API_URL: 'http://127.0.0.1:8080',
    },
    specPattern: 'cypress/e2e/**/*.cy.{js,ts}',
    supportFile: 'cypress/support/e2e.js',
    video: false,
    screenshotOnRunFailure: false,
  },
});

