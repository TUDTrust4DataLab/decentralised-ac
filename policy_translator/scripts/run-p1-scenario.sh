#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SPEC="$ROOT/normative/scenarios/p1/p1-combined.eflint"
SCENARIOS="$ROOT/normative/scenarios/p1"
RESULTS="$ROOT/results/smoke/p1"

if ! command -v eflint-repl >/dev/null 2>&1; then
    echo "Error: eflint-repl not found in PATH"
    exit 1
fi

mkdir -p "$RESULTS"

# Rebuild the combined specification.
{
    cat "$ROOT/normative/specifications/common.eflint"
    printf '\n\n'
    cat "$ROOT/normative/specifications/p1-admission.eflint"
    printf '\n'
} > "$SPEC"

run_scenario() {
    local name="$1"
    local input="$SCENARIOS/$name.txt"
    local output="$RESULTS/$name.log"

    echo "Running: $name"

    eflint-repl "$SPEC" < "$input" > "$output" 2>&1

    echo "Output saved: $output"
}


for scenario in \
    positive \
    missing-approval \
    missing-terms \
    wrong-purpose \
    wrong-session \
    unauthorised-approval
do
    run_scenario "$scenario"
done


echo "P1 scenarios finished."
echo "Review logs in: $RESULTS"
