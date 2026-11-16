import os
import subprocess
import sys
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[1]


def test_fabric_mode_does_not_load_solidity_dependencies():
    code = r'''
import builtins
import os

os.environ["ACCESS_CONTROL_BACKEND"] = "fabric"
os.environ["FABRIC_GATEWAY_URL"] = "http://localhost:5524"

real_import = builtins.__import__


def guarded_import(
    name,
    globals=None,
    locals=None,
    fromlist=(),
    level=0,
):
    blocked = (
        "solcx",
        "blockchain_interface.interfaces.ABACContracts",
        "blockchain_interface.interfaces.HyperledgerBesu",
    )

    if name.startswith(blocked):
        raise AssertionError(
            f"EVM/Solidity dependency loaded "
            f"in Fabric mode: {name}"
        )

    return real_import(
        name,
        globals,
        locals,
        fromlist,
        level,
    )


builtins.__import__ = guarded_import

import api

print("Fabric startup import successful")
'''

    env = os.environ.copy()
    env['ACCESS_CONTROL_BACKEND'] = 'fabric'
    env['FABRIC_GATEWAY_URL'] = (
        'http://localhost:5524'
    )

    result = subprocess.run(
        [
            sys.executable,
            '-c',
            code,
        ],
        cwd=PROJECT_ROOT,
        env=env,
        capture_output=True,
        text=True,
    )

    assert result.returncode == 0, (
        f'Application failed to import in Fabric mode.\n'
        f'STDOUT:\n{result.stdout}\n'
        f'STDERR:\n{result.stderr}'
    )

    assert (
        'Fabric startup import successful'
        in result.stdout
    )