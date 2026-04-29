#!/usr/bin/env python3
import argparse
import json
import sys
from pathlib import Path


def read_json(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def validate(security: dict, reliability: dict, performance: dict) -> tuple[bool, list[str]]:
    errors: list[str] = []

    if security.get("backend_tests_passed") is not True:
        errors.append("security gate failed: backend_tests_passed must be true")
    if security.get("frontend_tests_passed") is not True:
        errors.append("security gate failed: frontend_tests_passed must be true")
    if security.get("dependency_audit_passed") is not True:
        errors.append("security gate failed: dependency_audit_passed must be true")

    if reliability.get("chaos_scenarios_total", 0) < 2:
        errors.append("reliability gate failed: require >= 2 chaos scenarios")
    if reliability.get("chaos_scenarios_passed", 0) < reliability.get("chaos_scenarios_total", 0):
        errors.append("reliability gate failed: all chaos scenarios must pass")

    if performance.get("soak_duration_minutes", 0) < 5:
        errors.append("performance gate failed: soak_duration_minutes must be >= 5")
    if performance.get("p95_latency_ms", 10_000) > 750:
        errors.append("performance gate failed: p95_latency_ms must be <= 750")
    if performance.get("error_rate_percent", 100) > 1.0:
        errors.append("performance gate failed: error_rate_percent must be <= 1.0")

    return len(errors) == 0, errors


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate clinical-grade release gates evidence.")
    parser.add_argument("--security", required=True, help="Path to security evidence JSON")
    parser.add_argument("--reliability", required=True, help="Path to reliability evidence JSON")
    parser.add_argument("--performance", required=True, help="Path to performance evidence JSON")
    parser.add_argument("--summary", required=True, help="Output path for gate summary JSON")
    args = parser.parse_args()

    security = read_json(Path(args.security))
    reliability = read_json(Path(args.reliability))
    performance = read_json(Path(args.performance))
    passed, errors = validate(security, reliability, performance)

    summary = {
        "release_gate_passed": passed,
        "security_gate_passed": security.get("backend_tests_passed") is True
        and security.get("frontend_tests_passed") is True
        and security.get("dependency_audit_passed") is True,
        "reliability_gate_passed": reliability.get("chaos_scenarios_total", 0) >= 2
        and reliability.get("chaos_scenarios_passed", 0) >= reliability.get("chaos_scenarios_total", 0),
        "performance_gate_passed": performance.get("soak_duration_minutes", 0) >= 5
        and performance.get("p95_latency_ms", 10_000) <= 750
        and performance.get("error_rate_percent", 100) <= 1.0,
        "errors": errors,
    }

    summary_path = Path(args.summary)
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    with summary_path.open("w", encoding="utf-8") as f:
        json.dump(summary, f, indent=2)

    if not passed:
        print("Clinical release gates failed:")
        for err in errors:
            print(f"- {err}")
        return 1

    print("Clinical release gates passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())

