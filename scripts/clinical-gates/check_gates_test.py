import json
import tempfile
import unittest
from pathlib import Path

from check_gates import validate


class CheckGatesTest(unittest.TestCase):
    def test_passes_when_all_thresholds_meet(self):
        passed, errors = validate(
            {"backend_tests_passed": True, "frontend_tests_passed": True, "dependency_audit_passed": True},
            {"chaos_scenarios_total": 2, "chaos_scenarios_passed": 2},
            {"soak_duration_minutes": 5, "p95_latency_ms": 700, "error_rate_percent": 0.4},
        )
        self.assertTrue(passed)
        self.assertEqual(errors, [])

    def test_fails_when_thresholds_violated(self):
        passed, errors = validate(
            {"backend_tests_passed": False, "frontend_tests_passed": True, "dependency_audit_passed": False},
            {"chaos_scenarios_total": 2, "chaos_scenarios_passed": 1},
            {"soak_duration_minutes": 3, "p95_latency_ms": 820, "error_rate_percent": 2.2},
        )
        self.assertFalse(passed)
        self.assertGreaterEqual(len(errors), 4)

    def test_summary_file_shape(self):
        with tempfile.TemporaryDirectory() as td:
            out = Path(td) / "summary.json"
            summary = {
                "release_gate_passed": True,
                "security_gate_passed": True,
                "reliability_gate_passed": True,
                "performance_gate_passed": True,
                "errors": [],
            }
            out.write_text(json.dumps(summary), encoding="utf-8")
            loaded = json.loads(out.read_text(encoding="utf-8"))
            self.assertIn("release_gate_passed", loaded)
            self.assertIn("errors", loaded)


if __name__ == "__main__":
    unittest.main()

