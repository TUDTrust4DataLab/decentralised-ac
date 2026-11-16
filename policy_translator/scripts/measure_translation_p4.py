
#!/usr/bin/env python3

import csv
import hashlib
import json
import statistics
import sys
import time
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from src.validator.eflint_validator import EFlintValidator
from src.structure_builder.builder import build_nir
from src.policy_builder.fabric.builder import build_chaincode

SPEC = ROOT / "normative/scenarios/p4/p4-combined.eflint"
OUTPUT = ROOT / "results/experiments/e3/p4"
RUNS = 500
WARMUP_RUNS = 10

METRICS = [
    "validation_ms",
    "nir_ms",
    "generation_ms",
    "total_ms",
]


def main():
    source = SPEC.read_text(encoding="utf-8")
    source_hash = hashlib.sha256(
        source.encode("utf-8")
    ).hexdigest()

    validator = EFlintValidator()
    OUTPUT.mkdir(parents=True, exist_ok=True)

    results = []

    for run in range(1 - WARMUP_RUNS, RUNS + 1):
        start = time.perf_counter_ns()

        validation = validator.validate(SPEC)
        if not validation.valid:
            raise RuntimeError(
                f"Validation failed in run {run}: "
                f"{validation.diagnostics}"
            )

        nir, nir_ns = build_nir(source)
        code, generation_ns = build_chaincode(nir)

        total_ns = time.perf_counter_ns() - start

        if not code.strip():
            raise RuntimeError("Empty generated chaincode")

        result = {
            "run_id": run,
            "validation_ms": validation.duration_ns / 1e6,
            "nir_ms": nir_ns / 1e6,
            "generation_ms": generation_ns / 1e6,
            "total_ms": total_ns / 1e6,
        }

        if run <= 0:
            continue  # Warm-up: excluded from results

        results.append(result)
        print(
            f"Run {run:02d}/{RUNS}: "
            f"{result['total_ms']:.3f} ms"
        )

    with (OUTPUT / "runs.csv").open(
        "w", newline=""
    ) as file:
        writer = csv.DictWriter(
            file,
            fieldnames=["run_id", *METRICS],
        )
        writer.writeheader()
        writer.writerows(results)

    summary = {
        "experiment": "E3",
        "policy": "P1",
        "runs": RUNS,
        "timestamp_utc": datetime.now(
            timezone.utc
        ).isoformat(),
        "specification_sha256": source_hash,
        "metrics": {},
    }

    for metric in METRICS:
        values = [row[metric] for row in results]

        summary["metrics"][metric] = {
            "mean": statistics.mean(values),
            "std": statistics.stdev(values),
            "min": min(values),
            "max": max(values),
            "median": statistics.median(values),
        }

    (OUTPUT / "summary.json").write_text(
        json.dumps(summary, indent=2)
    )

    print("\n=== SUMMARY (milliseconds) ===")
    for metric, stats in summary["metrics"].items():
        print(
            f"{metric}: "
            f"{stats['mean']:.3f} ± "
            f"{stats['std']:.3f}"
        )

    print(f"\nSaved to: {OUTPUT}")


if __name__ == "__main__":
    main()
