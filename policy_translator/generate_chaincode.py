import argparse
import json
from pathlib import Path

from src.policy_builder.fabric.builder import generate


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("nir", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    result = generate(args.nir, args.output)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
