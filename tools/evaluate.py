#!/usr/bin/env python3
"""Run a small, private semantic-validation evaluation against the REST API."""

import argparse
import json
import os
import re
import sys
import urllib.error
import urllib.request
from collections import Counter
from pathlib import Path
from urllib.parse import urlsplit


class NoRedirect(urllib.request.HTTPRedirectHandler):
    """Never forward an API key to a URL supplied by a redirect response."""

    def redirect_request(self, request, fp, code, msg, headers, new_url):
        return None


def open_request(request, timeout):
    return urllib.request.build_opener(NoRedirect).open(request, timeout=timeout)


def load_cases(path):
    data = json.loads(Path(path).read_text(encoding="utf-8"))
    if not isinstance(data, list) or not data:
        raise ValueError("cases must be a nonempty JSON array")
    seen = set()
    for index, case in enumerate(data, 1):
        if not isinstance(case, dict):
            raise ValueError(f"case {index}: expected an object")
        case_id = case.get("id")
        if not isinstance(case_id, str) or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_-]{0,79}", case_id) or case_id in seen:
            raise ValueError(f"case {index}: id must be unique and contain only letters, digits, _ or -")
        seen.add(case_id)
        if type(case.get("expected_valid")) is not bool:
            raise ValueError(f"case {index}: expected_valid must be true or false")
        value = case.get("value")
        if not isinstance(value, str) or not 1 <= len(value.strip()) <= 10000:
            raise ValueError(f"case {index}: value must contain 1 to 10000 characters")
        has_rule = "rule" in case
        has_question = "question" in case
        if has_rule == has_question:
            raise ValueError(f"case {index}: supply exactly one of rule or question")
        for field in ("rule", "question"):
            if field in case and (not isinstance(case[field], str) or not case[field].strip()):
                raise ValueError(f"case {index}: {field} must be a nonempty string")
    return data


def check_url(base_url):
    parsed = urlsplit(base_url)
    if parsed.scheme not in ("https", "http") or not parsed.hostname:
        raise ValueError("base URL must be an absolute HTTP(S) URL")
    if parsed.scheme == "http" and parsed.hostname not in ("localhost", "127.0.0.1", "::1"):
        raise ValueError("HTTP is allowed only for local development")
    if parsed.username or parsed.password or parsed.query or parsed.fragment or parsed.path not in ("", "/"):
        raise ValueError("base URL must not contain credentials, path, query or fragment")
    return base_url.rstrip("/")


def evaluate_case(base_url, api_key, case, timeout):
    endpoint = "/v1/validate" if "rule" in case else "/v1/check"
    payload = {"value": case["value"]}
    payload["rule" if "rule" in case else "question"] = case.get("rule", case.get("question"))
    request = urllib.request.Request(
        base_url + endpoint,
        data=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
        headers={"Authorization": "Bearer " + api_key, "Content-Type": "application/json"},
        method="POST",
    )
    with open_request(request, timeout=timeout) as response:
        result = json.load(response)
    if not isinstance(result, dict):
        raise ValueError("invalid API response")
    valid, status, confidence = result.get("valid"), result.get("status"), result.get("confidence")
    if type(confidence) not in (int, float) or not 0 <= confidence <= 1:
        raise ValueError("invalid API response")
    if not ((valid is True and status == "valid") or
            (valid is False and status == "invalid") or
            (valid is None and status == "uncertain")):
        raise ValueError("invalid API response")
    return valid, status, confidence


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cases", default="evals/cases.example.json")
    parser.add_argument("--base-url", default="http://localhost:8080")
    parser.add_argument("--limit", type=int, default=100)
    parser.add_argument("--timeout", type=float, default=40)
    parser.add_argument("--dry-run", action="store_true", help="validate cases without making API requests")
    args = parser.parse_args(argv)
    try:
        cases = load_cases(args.cases)
        base_url = check_url(args.base_url)
        if not 1 <= args.limit <= 100 or args.timeout <= 0:
            raise ValueError("limit must be 1..100 and timeout must be positive")
    except (ValueError, OSError, json.JSONDecodeError) as exc:
        print(f"Invalid evaluation setup: {exc}", file=sys.stderr)
        return 2
    selected = cases[:args.limit]
    if args.dry_run:
        print(f"Ready: {len(selected)} cases; no API requests sent")
        return 0
    api_key = os.getenv("SEMANTIC_VALIDATOR_API_KEY", "")
    if not api_key.strip():
        print("SEMANTIC_VALIDATOR_API_KEY is required", file=sys.stderr)
        return 2
    counts = Counter()
    for case in selected:
        try:
            valid, status, confidence = evaluate_case(base_url, api_key, case, args.timeout)
        except urllib.error.HTTPError as exc:
            counts["errors"] += 1
            print(f"{case['id']}: HTTP {exc.code}")
            if exc.code == 429:
                print("Quota or rate limit reached; stopping")
                break
            continue
        except (urllib.error.URLError, TimeoutError, ValueError, json.JSONDecodeError):
            counts["errors"] += 1
            print(f"{case['id']}: request or response error")
            continue
        expected = case["expected_valid"]
        if valid is None:
            outcome = "uncertain"
        elif valid == expected:
            outcome = "correct"
        elif valid:
            outcome = "false_positive"
        else:
            outcome = "false_negative"
        counts[outcome] += 1
        print(f"{case['id']}: expected={str(expected).lower()} actual={status} confidence={confidence:.2f} outcome={outcome}")
    print("Summary: " + ", ".join(f"{key}={counts[key]}" for key in
                              ("correct", "false_positive", "false_negative", "uncertain", "errors")))
    print(f"Completed {sum(counts.values())}/{len(selected)} cases; each request may consume quota")
    return 1 if counts["errors"] else 0


if __name__ == "__main__":
    raise SystemExit(main())
