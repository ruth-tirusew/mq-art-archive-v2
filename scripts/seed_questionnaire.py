#!/usr/bin/env python3
"""Load a questionnaire definition (JSON) into an environment as a DRAFT, through the admin API.

It goes through the same endpoints the admin editor uses, so the server validates it exactly as it
would validate an edit. It never publishes unless you pass --publish: by default an admin reviews the
draft in Admin -> Applications -> Questionnaire and presses Publish.

    SEED_ADMIN_PASSWORD=... python3 scripts/seed_questionnaire.py \\
        --api https://api.example.com --origin https://admin.example.com --email you@example.com

--origin must be one of the API's allowed CORS origins (the admin app's address): the API rejects
state-changing requests from any other origin, which is its CSRF protection.
"""

from __future__ import annotations

import argparse
import http.cookiejar
import json
import os
import pathlib
import sys
import urllib.error
import urllib.request

DEFAULT_FILE = pathlib.Path(__file__).resolve().parents[1] / "docs" / "questionnaire" / "artiv-archive-v1.1.json"


class Api:
    def __init__(self, base: str, origin: str):
        self.base, self.origin = base.rstrip("/"), origin
        self.opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))

    def call(self, method: str, path: str, body=None, *, allow=()):
        data = None if body is None else json.dumps(body).encode()
        req = urllib.request.Request(self.base + path, data=data, method=method)
        req.add_header("Origin", self.origin)
        if data is not None:
            req.add_header("Content-Type", "application/json")
        try:
            with self.opener.open(req, timeout=30) as resp:
                raw = resp.read()
                return resp.status, (json.loads(raw) if raw else None)
        except urllib.error.HTTPError as err:
            raw = err.read()
            payload = json.loads(raw) if raw else None
            if err.code in allow:
                return err.code, payload
            message = (payload or {}).get("error", raw.decode(errors="replace")) if isinstance(payload, dict) else raw.decode(errors="replace")
            sys.exit(f"{method} {path} failed: {err.code} {message}")
        except urllib.error.URLError as err:
            sys.exit(f"cannot reach {self.base}: {err.reason}")


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--api", default="http://localhost:8080", help="API base URL")
    ap.add_argument("--origin", default="http://localhost:5174", help="an allowed CORS origin, e.g. the admin app's address")
    ap.add_argument("--email", required=True, help="admin account email")
    ap.add_argument("--file", type=pathlib.Path, default=DEFAULT_FILE)
    ap.add_argument("--replace-draft", action="store_true", help="overwrite an existing non-empty draft (otherwise refuse)")
    ap.add_argument("--publish", action="store_true", help="also publish the draft as the next revision (default: leave it as a draft)")
    args = ap.parse_args()

    password = os.environ.get("SEED_ADMIN_PASSWORD")
    if not password:
        sys.exit("set SEED_ADMIN_PASSWORD in the environment (kept out of the command line and shell history)")
    payload = json.loads(args.file.read_text(encoding="utf-8"))

    api = Api(args.api, args.origin)
    api.call("POST", "/api/v1/auth/login", {"email": args.email, "password": password})

    status, draft = api.call("GET", "/admin/v1/questionnaire/draft", allow=(404,))
    if status == 200 and draft.get("questions") and not args.replace_draft:
        sys.exit(f"there is already a draft with {len(draft['questions'])} questions; "
                 "review or discard it in the admin app, or pass --replace-draft to overwrite it")

    _, saved = api.call("PUT", "/admin/v1/questionnaire/draft", payload)
    print(f"saved draft: {len(saved['questions'])} questions from {args.file.name}")

    if args.publish:
        _, rev = api.call("POST", "/admin/v1/questionnaire/draft/publish")
        print(f"published as revision {rev['number']}")
    else:
        print("left as a draft. Review it in the admin app (Applications -> Questionnaire) and publish when ready.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
