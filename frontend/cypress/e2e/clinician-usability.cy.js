// CGS-03 — clinician usability benchmark (repeatable, seeded-data flows)
describe("Clinician usability benchmarks (CGS-03)", () => {
  const API = Cypress.env("API_URL");

  const thresholdMs = {
    aiEval: 30000,
    ackEscalation: 15000,
    patientDashboardLoad: 15000,
  };

  it("admin: AI eval time-to-task (Run eval -> Eval result)", () => {
    const start = Date.now();

    cy.request("POST", `${API}/api/login`, {
      email: "admin@healthonyx.demo",
      password: "admin123",
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit("/admin/ai", {
        onBeforeLoad(win) {
          win.localStorage.setItem("token", token);
          win.localStorage.setItem("user", JSON.stringify(user));
        },
      });
    });

    cy.get("[data-cy='ai-runtime-health']", { timeout: 20000 }).should("exist");
    cy.contains("Run eval").click();
    cy.contains("Eval result", { timeout: 40000 }).should("exist");

    cy.then(() => {
      const elapsed = Date.now() - start;
      cy.log(`AI eval time-to-task: ${elapsed} ms`);
      expect(elapsed).to.be.lessThan(thresholdMs.aiEval);
    });
  });

  it("doctor: critical escalation acknowledgement time-to-task", () => {
    const start = Date.now();

    cy.request("POST", `${API}/api/login`, {
      email: "doctor@healthonyx.demo",
      password: "doctor123",
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit("/doctor/dashboard", {
        onBeforeLoad(win) {
          win.localStorage.setItem("token", token);
          win.localStorage.setItem("user", JSON.stringify(user));
        },
      });
    });

    cy.get("[data-cy='doctor-critical-escalations']", { timeout: 30000 }).should("exist");
    cy.get("button[data-cy^='ack-critical-']")
      .should("not.be.disabled")
      .first()
      .click();

    // After acknowledgement, the corresponding button becomes disabled.
    cy.get("button[data-cy^='ack-critical-']")
      .first()
      .should("be.disabled");

    cy.then(() => {
      const elapsed = Date.now() - start;
      cy.log(`Critical escalation ack time-to-task: ${elapsed} ms`);
      expect(elapsed).to.be.lessThan(thresholdMs.ackEscalation);
    });
  });

  it("patient: dashboard load time-to-task", () => {
    const start = Date.now();

    cy.request("POST", `${API}/api/login`, {
      email: "patient@healthonyx.demo",
      password: "patient123",
    }).then((res) => {
      const { token, user } = res.body;
      cy.visit("/patient/dashboard", {
        onBeforeLoad(win) {
          win.localStorage.setItem("token", token);
          win.localStorage.setItem("user", JSON.stringify(user));
        },
      });
    });

    cy.get('[data-cy="patient-dashboard"]', { timeout: 20000 }).should("be.visible");

    cy.then(() => {
      const elapsed = Date.now() - start;
      cy.log(`Patient dashboard load time-to-task: ${elapsed} ms`);
      expect(elapsed).to.be.lessThan(thresholdMs.patientDashboardLoad);
    });
  });
});

