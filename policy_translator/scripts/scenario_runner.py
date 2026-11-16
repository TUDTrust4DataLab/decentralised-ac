#!/usr/bin/env python3

import json
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SPEC_DIR = ROOT / "normative/specifications"
SCENARIO_DIR = ROOT / "normative/scenarios/p1"
RESULTS_DIR = ROOT / "results/smoke/p1"

FAILED_QUERY = re.compile(r"\bquery failed\b", re.I)
ERROR_PATTERN = re.compile(
    r"could not parse|unsuccessful parse|"
    r"type error|unknown command|"
    r"not enabled|not permitted",
    re.I,
)


def build_specification():
    combined = SCENARIO_DIR / "p1-combined.eflint"
    combined.write_text(
        (SPEC_DIR / "common.eflint").read_text()
        + "\n\n"
        + (SPEC_DIR / "p1-admission.eflint").read_text()
        + "\n"
    )
    return combined


def extract_checkpoints(scenario_file):
    history = []
    checkpoints = []

    for line in scenario_file.read_text().splitlines():
        command = line.strip()

        if not command or command.startswith("#"):
            continue

        if command.startswith("?"):
            checkpoints.append({
                "query": command,
                "history": list(history),
            })
        elif not command.startswith(":"):
            history.append(command)

    return checkpoints


def run_checkpoint(specification, checkpoint):
    commands = (
        checkpoint["history"]
        + [checkpoint["query"], ":quit"]
    )

    process = subprocess.run(
        ["eflint-repl", str(specification)],
        input="\n".join(commands) + "\n",
        text=True,
        capture_output=True,
        timeout=60,
        check=False,
    )

    output = process.stdout + "\n" + process.stderr

    if process.returncode != 0:
        raise RuntimeError(
            f"Interpreter exit code: {process.returncode}"
        )

    if ERROR_PATTERN.search(output):
        raise RuntimeError(
            "Interpreter reported an execution error"
        )

    return not bool(FAILED_QUERY.search(output)), output


def main():
    manifest = json.loads(
        (SCENARIO_DIR / "expected.json").read_text()
    )
    specification = build_specification()

    RESULTS_DIR.mkdir(parents=True, exist_ok=True)
    results = []

    for scenario in manifest["scenarios"]:
        checkpoints = extract_checkpoints(
            SCENARIO_DIR / scenario["file"]
        )
        expected = scenario["expected_queries"]

        if len(checkpoints) != len(expected):
            raise ValueError(
                f'{scenario["id"]}: expected '
                f'{len(expected)} queries, found '
                f'{len(checkpoints)}'
            )

        assertions = []

        for index, (checkpoint, expectation) in enumerate(
            zip(checkpoints, expected), start=1
        ):
            try:
                observed, output = run_checkpoint(
                    specification, checkpoint
                )
                status = (
                    "PASS" if observed == expectation
                    else "FAIL"
                )
                error = None
            except (OSError, subprocess.TimeoutExpired,
                    RuntimeError) as exc:
                observed = None
                output = str(exc)
                status = "ERROR"
                error = str(exc)

            log_name = f'{scenario["id"]}-Q{index:02d}.log'
            (RESULTS_DIR / log_name).write_text(output)

            assertions.append({
                "query": checkpoint["query"],
                "expected": expectation,
                "observed": observed,
                "status": status,
                "error": error,
            })

        scenario_status = (
            "PASS"
            if all(a["status"] == "PASS"
                   for a in assertions)
            else "FAIL"
        )

        results.append({
            "test_id": scenario["id"],
            "status": scenario_status,
            "assertions": assertions,
        })

        print(
            f'{scenario["id"]}: {scenario_status} '
            f'({sum(a["status"] == "PASS" for a in assertions)}'
            f'/{len(assertions)} assertions)'
        )

    report = {
        "policy": "P1",
        "timestamp": datetime.now(
            timezone.utc
        ).isoformat(),
        "results": results,
    }

    (RESULTS_DIR / "report.json").write_text(
        json.dumps(report, indent=2)
    )

    passed = sum(
        result["status"] == "PASS"
        for result in results
    )

    print(f"\nResults: {passed}/{len(results)} passed")

    return 0 if passed == len(results) else 1


if __name__ == "__main__":
    sys.exit(main())
