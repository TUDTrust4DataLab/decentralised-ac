import argparse
import hashlib
import json
from pathlib import Path

from src.structure_builder.builder import build_nir


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("specification", type=Path)
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("generated/nir/p1.nir.json"),
    )
    args = parser.parse_args()

    source = args.specification.read_text(
        encoding="utf-8"
    )

    nir, duration_ns = build_nir(source)

    nir["source"] = {
        "file": str(args.specification),
        "sha256": hashlib.sha256(
            source.encode("utf-8")
        ).hexdigest(),
    }

    args.output.parent.mkdir(
        parents=True, exist_ok=True
    )

    args.output.write_text(
        json.dumps(nir, indent=2),
        encoding="utf-8",
    )

    print(json.dumps({
        "status": "success",
        "stage": "nir_construction",
        "duration_ns": duration_ns,
        "facts": len(nir["facts"]),
        "acts": len(nir["acts"]),
        "placeholders": len(nir["placeholders"]),
        "output": str(args.output),
    }, indent=2))


if __name__ == "__main__":
    main()
