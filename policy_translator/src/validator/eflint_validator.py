import re
import shutil
import subprocess
import time

from dataclasses import dataclass
from pathlib import Path


ERROR_PATTERN = re.compile(
    r"could not parse flint|"
    r"could not parse flint spec|"
    r"unsuccessful parse|"
    r"type error|"
    r"unknown command",
    re.IGNORECASE,
)


@dataclass
class ValidationResult:
    valid: bool
    duration_ns: int
    diagnostics: str


class EFlintValidator:

    def __init__(self, executable: str = "eflint-repl"):
        self.executable = shutil.which(executable)

        if self.executable is None:
            raise RuntimeError(
                f"eFLINT executable not found: {executable}"
            )

    def validate(self, specification: Path) -> ValidationResult:
        if not specification.is_file():
            raise FileNotFoundError(specification)

        start = time.perf_counter_ns()

        try:
            result = subprocess.run(
                [self.executable, str(specification.resolve())],
                input=":quit\n",
                text=True,
                capture_output=True,
                timeout=60,
                check=False,
            )

            diagnostics = (
                result.stdout + "\n" + result.stderr
            ).strip()

            valid = (
                result.returncode == 0
                and not ERROR_PATTERN.search(diagnostics)
            )

        except subprocess.TimeoutExpired:
            valid = False
            diagnostics = "Validation timed out"

        duration_ns = time.perf_counter_ns() - start

        return ValidationResult(
            valid=valid,
            duration_ns=duration_ns,
            diagnostics=diagnostics,
        )
