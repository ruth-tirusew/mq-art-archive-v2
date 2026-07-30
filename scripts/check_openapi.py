#!/usr/bin/env python3
"""Parse and lightly validate contracts/openapi.yaml."""

from __future__ import annotations

import pathlib
import sys

try:
    import yaml
except ImportError:
    import subprocess

    subprocess.check_call([sys.executable, "-m", "pip", "install", "pyyaml", "-q"])
    import yaml


def main() -> int:
    root = pathlib.Path(__file__).resolve().parents[1]
    path = root / "contracts" / "openapi.yaml"
    doc = yaml.safe_load(path.read_text())
    if not isinstance(doc, dict):
        print("openapi: document is not a mapping", file=sys.stderr)
        return 1
    if not doc.get("openapi"):
        print("openapi: missing openapi version", file=sys.stderr)
        return 1
    paths = doc.get("paths")
    if not isinstance(paths, dict) or not paths:
        print("openapi: missing paths", file=sys.stderr)
        return 1
    print(f"openapi ok: version={doc['openapi']} paths={len(paths)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
