
#!/usr/bin/env python3

import csv
import json
import math
import statistics
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
RESULTS = ROOT / "results/experiments/e3"

POLICIES = ["p1", "p2", "p3", "p4", "p5"]

METRICS = [
    "validation_ms",
    "nir_ms",
    "generation_ms",
    "total_ms",
]


def summarise(values):
    n = len(values)

    if n < 2:
        raise ValueError(
            "At least two observations are required"
        )

    mean = statistics.mean(values)
    std = statistics.stdev(values)

    # Normal approximation for the mean's 95% CI.
    margin = 1.96 * std / math.sqrt(n)

    ordered = sorted(values)

    def percentile(p):
        index = (n - 1) * p
        lower = math.floor(index)
        upper = math.ceil(index)
        weight = index - lower

        return (
            ordered[lower] * (1 - weight)
            + ordered[upper] * weight
        )

    return {
        "n": n,
        "mean": mean,
        "std": std,
        "median": statistics.median(values),
        "p95": percentile(0.95),
        "p99": percentile(0.99),
        "ci95_low": mean - margin,
        "ci95_high": mean + margin,
        "ci95_half_width": margin,
        "relative_ci95_percent": (
            100 * margin / mean
            if mean != 0 else None
        ),
    }


def get_complexity(policy):
    spec = (
        ROOT / "normative/scenarios"
        / policy / f"{policy}-combined.eflint"
    )

    nir_path = (
        ROOT / "generated/nir"
        / f"{policy}.nir.json"
    )

    if not spec.exists() or not nir_path.exists():
        raise FileNotFoundError(
            f"Missing specification or NIR: {policy}"
        )

    source = spec.read_text(encoding="utf-8")
    nir = json.loads(nir_path.read_text())

    return {
        "source_lines": len(source.splitlines()),
        "source_bytes": len(source.encode("utf-8")),
        "facts": len(nir.get("facts", [])),
        "duties": len(nir.get("duties", [])),
        "acts": len(nir.get("acts", [])),
    }


def main():
    comparison = []
    detailed = {}
    missing = []

    for policy in POLICIES:
        runs_file = RESULTS / policy / "runs.csv"

        if not runs_file.exists():
            missing.append(policy)
            continue

        with runs_file.open(newline="") as file:
            runs = list(csv.DictReader(file))

        if not runs:
            raise ValueError(
                f"No observations for {policy}"
            )

        metrics = {}

        for metric in METRICS:
            values = [
                float(row[metric]) for row in runs
            ]

            if not all(
                math.isfinite(x) and x >= 0
                for x in values
            ):
                raise ValueError(
                    f"Invalid values: {policy}/{metric}"
                )

            metrics[metric] = summarise(values)

        complexity = get_complexity(policy)

        detailed[policy] = {
            "complexity": complexity,
            "metrics": metrics,
        }

        total = metrics["total_ms"]

        comparison.append({
            "policy": policy.upper(),
            "runs": total["n"],
            **complexity,
            "mean_total_ms": total["mean"],
            "std_total_ms": total["std"],
            "median_total_ms": total["median"],
            "p95_total_ms": total["p95"],
            "p99_total_ms": total["p99"],
            "ci95_low_ms": total["ci95_low"],
            "ci95_high_ms": total["ci95_high"],
            "relative_ci95_percent":
                total["relative_ci95_percent"],
            "mean_validation_ms":
                metrics["validation_ms"]["mean"],
            "mean_nir_ms":
                metrics["nir_ms"]["mean"],
            "mean_generation_ms":
                metrics["generation_ms"]["mean"],
        })

    if not comparison:
        raise RuntimeError("No experiment data found")

    RESULTS.mkdir(parents=True, exist_ok=True)

    output_csv = RESULTS / "comparison.csv"

    with output_csv.open(
        "w", newline=""
    ) as file:
        writer = csv.DictWriter(
            file,
            fieldnames=list(comparison[0].keys()),
        )
        writer.writeheader()
        writer.writerows(comparison)

    output_json = RESULTS / "comparison.json"

    output_json.write_text(
        json.dumps({
            "policies": detailed,
            "missing_policies": missing,
            "confidence_interval_method":
                "Normal approximation (1.96 * SE)",
        }, indent=2)
    )

    print("\n=== E3: TRANSLATION PERFORMANCE ===\n")

    for row in comparison:
        print(
            f"{row['policy']}: "
            f"n={row['runs']}, "
            f"mean={row['mean_total_ms']:.3f} ms, "
            f"std={row['std_total_ms']:.3f} ms, "
            f"95% CI=[{row['ci95_low_ms']:.3f}, "
            f"{row['ci95_high_ms']:.3f}] ms"
        )

    if missing:
        print(
            "\nMissing experiments: "
            + ", ".join(missing)
        )

    print(f"\nSaved: {output_csv}")
    print(f"Saved: {output_json}")


if __name__ == "__main__":
    main()
