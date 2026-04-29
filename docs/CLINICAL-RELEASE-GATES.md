# Clinical Release Gates

Healthonyx uses mandatory release gates to block shipping when security, reliability, or performance criteria are not met.

## Gates enforced in CI

The `Clinical Release Gates` job in `.github/workflows/ci.yml` runs after backend and frontend jobs.

- **Security gate**
  - Backend security-focused test subset
  - Frontend test suite execution
  - Security evidence JSON generated
- **Reliability/chaos gate**
  - Callback/reliability incident test subset
  - Chaos evidence JSON generated
- **Performance gate**
  - SLO/performance baseline test subset
  - Performance evidence JSON generated

The checker `scripts/clinical-gates/check_gates.py` validates threshold criteria and fails CI when any gate is below policy.

## Evidence artifacts

Each run publishes the artifact `clinical-release-gate-evidence` containing:

- `security-evidence.json`
- `reliability-evidence.json`
- `performance-evidence.json`
- `release-gate-summary.json`

## Pass criteria

- `backend_tests_passed == true`
- `frontend_tests_passed == true`
- `dependency_audit_passed == true`
- `chaos_scenarios_total >= 2`
- `chaos_scenarios_passed == chaos_scenarios_total`
- `soak_duration_minutes >= 5`
- `p95_latency_ms <= 750`
- `error_rate_percent <= 1.0`

If any condition fails, `release_gate_summary.release_gate_passed` is `false` and the workflow fails.

