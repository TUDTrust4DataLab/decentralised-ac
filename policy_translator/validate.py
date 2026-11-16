import argparse
import json
from pathlib import Path

from src.validator.eflint_validator import EFlintValidator


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("specification", type=Path)
    args = parser.parse_args()

    validator = EFlintValidator()
    result = validator.validate(args.specification)

    output = {
        "stage": "validation",
        "status": "success" if result.valid else "error",
        "duration_ns": result.duration_ns,
        "diagnostics": result.diagnostics,
    }

    print(json.dumps(output, indent=2))

    return 0 if result.valid else 1


if __name__ == "__main__":
    raise SystemExit(main())
