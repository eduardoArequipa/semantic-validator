#!/usr/bin/env python3
"""Compare the current hosted API with a local Laya server on labeled cases."""

import argparse
import json
import math
import os
import re
import statistics
import sys
import time
import urllib.error
import urllib.request
from collections import Counter, defaultdict
from pathlib import Path

import evaluate


RULE_QUESTIONS = {"person_name": "¿Este texto parece representar el nombre de una persona?"}  # v1
THRESHOLD = 0.80


def load_cases(path):
    groups = json.loads(Path(path).read_text(encoding="utf-8"))
    if not isinstance(groups, list) or not groups:
        raise ValueError("cases must be a nonempty array of groups")
    cases = []
    seen = set()
    for group in groups:
        if not isinstance(group, dict):
            raise ValueError("each group must be an object")
        scenario, language = group.get("scenario"), group.get("language")
        if not isinstance(scenario, str) or not re.fullmatch(r"[a-z][a-z0-9_-]{0,39}", scenario):
            raise ValueError("scenario must be a short safe identifier")
        if language not in ("es", "en") or (scenario, language) in seen:
            raise ValueError("language must be es/en and scenario-language pairs must be unique")
        seen.add((scenario, language))
        has_rule, has_question = "rule" in group, "question" in group
        if has_rule == has_question:
            raise ValueError("each group needs exactly one rule or question")
        if has_rule and (not isinstance(group["rule"], str) or group["rule"] not in RULE_QUESTIONS):
            raise ValueError("unknown local rule mapping")
        if has_question and (not isinstance(group["question"], str) or not group["question"].strip()
                             or len(group["question"]) > 10000):
            raise ValueError("question must contain 1 to 10000 characters")
        examples = group.get("examples")
        if not isinstance(examples, list) or not examples:
            raise ValueError("each group needs examples")
        for index, example in enumerate(examples, 1):
            if (not isinstance(example, list) or len(example) != 2 or
                    not isinstance(example[0], str) or not 1 <= len(example[0].strip()) <= 10000 or
                    type(example[1]) is not bool):
                raise ValueError("each example must be [nonempty text, boolean label]")
            case = {"id": f"{scenario}-{language}-{index:02d}", "scenario": scenario,
                    "language": language, "value": example[0], "expected_valid": example[1]}
            case["rule" if has_rule else "question"] = group["rule" if has_rule else "question"]
            cases.append(case)
    return cases


def laya_question(case):
    return case.get("question", RULE_QUESTIONS.get(case.get("rule")))


def parse_laya_response(document, primitive="noul"):
    if not isinstance(document, dict) or not isinstance(document.get("answers"), dict):
        raise ValueError("invalid Laya response")
    answer = document["answers"].get("result")
    if not isinstance(answer, dict):
        raise ValueError("invalid Laya response")
    if primitive == "choice":
        choice = answer.get("choice")
        confidence = answer.get("answer_confidence")  # Laya's `confidence` is entropy-based for choice.
        if choice not in ("A", "B") or type(confidence) not in (int, float) or \
                not math.isfinite(confidence) or not 0 <= confidence <= 1:
            raise ValueError("invalid Laya choice")
        valid = choice == "A"
    elif primitive == "noul":
        probability = answer.get("noul")
        if type(probability) not in (int, float) or not math.isfinite(probability) or not 0 <= probability <= 1:
            raise ValueError("invalid Laya probability")
        confidence = max(probability, 1 - probability)
        valid = probability >= 0.5
    else:
        raise ValueError("unknown Laya primitive")
    if confidence < THRESHOLD:
        return None, "uncertain", confidence
    return valid, "valid" if valid else "invalid", confidence


def evaluate_laya(base_url, api_key, case, timeout, primitive="noul"):
    if case["language"] == "es":
        criteria = {"true": "el texto cumple la condición", "false": "el texto no cumple la condición"}
    else:
        criteria = {"true": "the text satisfies the condition", "false": "the text does not satisfy the condition"}
    if primitive == "choice":
        question = {"type": "choice", "instructions": laya_question(case),
                    "criteria": {"A": criteria["true"], "B": criteria["false"]}}
    elif primitive == "noul":
        question = {"type": "noul", "instructions": laya_question(case),
                    "criteria": criteria, "labels": {"true": "A", "false": "B"}}
    else:
        raise ValueError("unknown Laya primitive")
    payload = {"state": case["value"], "questions": {"result": question}}
    headers = {"Content-Type": "application/json"}
    if api_key:
        headers["Authorization"] = "Bearer " + api_key
    request = urllib.request.Request(
        base_url + "/v1/systemone", data=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
        headers=headers, method="POST",
    )
    with evaluate.open_request(request, timeout) as response:
        return parse_laya_response(json.load(response), primitive)


def outcome(valid, expected):
    if valid is None:
        return "uncertain"
    if valid == expected:
        return "correct"
    return "false_positive" if valid else "false_negative"


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cases", default="evals/provider-comparison.example.json")
    parser.add_argument("--jev-base-url", default="https://validator.trialsur.cloud")
    parser.add_argument("--laya-base-url", default="http://127.0.0.1:8000")
    parser.add_argument("--limit", type=int, default=100)
    parser.add_argument("--timeout", type=float, default=40)
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--failures-only", action="store_true", help="print only non-correct case outcomes and the summary")
    parser.add_argument("--laya-only", action="store_true", help="do not call the hosted Jev API")
    parser.add_argument("--laya-primitive", choices=("noul", "choice"), default="noul")
    args = parser.parse_args(argv)
    try:
        cases = load_cases(args.cases)
        jev_url = evaluate.check_url(args.jev_base_url)
        laya_url = evaluate.check_url(args.laya_base_url)
        if not 1 <= args.limit <= 200 or args.timeout <= 0:
            raise ValueError("limit must be 1..200 and timeout must be positive")
    except (ValueError, OSError, json.JSONDecodeError) as exc:
        print(f"Invalid comparison setup: {exc}", file=sys.stderr)
        return 2
    cases = cases[:args.limit]
    if args.dry_run:
        request_count = len(cases) if args.laya_only else 2 * len(cases)
        print(f"Ready: {len(cases)} labeled cases; {request_count} requests maximum; none sent")
        return 0
    jev_key = os.getenv("SEMANTIC_VALIDATOR_API_KEY", "").strip()
    if not args.laya_only and not jev_key:
        print("SEMANTIC_VALIDATOR_API_KEY is required", file=sys.stderr)
        return 2
    laya_key = os.getenv("LAYA_API_KEY", "").strip()
    providers = ("laya",) if args.laya_only else ("jev", "laya")
    counts = {provider: Counter() for provider in providers}
    by_language = defaultdict(lambda: {provider: Counter() for provider in providers})
    by_scenario = defaultdict(lambda: {provider: Counter() for provider in providers})
    latencies = {provider: [] for provider in providers}
    paired = Counter()
    for case in cases:
        observed = {}
        calls = (("laya", lambda: evaluate_laya(laya_url, laya_key, case, args.timeout, args.laya_primitive)),)
        if not args.laya_only:
            calls = (("jev", lambda: evaluate.evaluate_case(jev_url, jev_key, case, args.timeout)),) + calls
        for provider, call in calls:
            start = time.perf_counter()
            try:
                valid, status, confidence = call()
            except urllib.error.HTTPError as exc:
                counts[provider]["errors"] += 1
                observed[provider] = "error"
                print(f"{case['id']} {provider}: HTTP {exc.code}")
                if provider == "jev" and exc.code == 429:
                    print("Jev quota or rate limit reached; stopping")
                    print_summary(counts, by_language, by_scenario, latencies, paired)
                    return 1
                continue
            except (urllib.error.URLError, TimeoutError, ValueError, json.JSONDecodeError):
                counts[provider]["errors"] += 1
                observed[provider] = "error"
                print(f"{case['id']} {provider}: request or response error")
                continue
            latencies[provider].append(time.perf_counter() - start)
            result = outcome(valid, case["expected_valid"])
            counts[provider][result] += 1
            by_language[case["language"]][provider][result] += 1
            by_scenario[case["scenario"]][provider][result] += 1
            observed[provider] = result
            if not args.failures_only or result != "correct":
                print(f"{case['id']} {provider}: expected={str(case['expected_valid']).lower()} "
                      f"actual={status} confidence={confidence:.2f} outcome={result}")
        if len(observed) == 2:
            paired[(observed["jev"], observed["laya"])] += 1
    print_summary(counts, by_language, by_scenario, latencies, paired)
    return 1 if any(counts[p]["errors"] for p in counts) else 0


def print_summary(counts, by_language, by_scenario, latencies, paired):
    fields = ("correct", "false_positive", "false_negative", "uncertain", "errors")
    providers = tuple(counts)
    for provider in providers:
        print(f"{provider}: " + ", ".join(f"{field}={counts[provider][field]}" for field in fields))
        if latencies[provider]:
            print(f"{provider} median observed request latency: {statistics.median(latencies[provider]) * 1000:.1f} ms")
    for title, groups in (("language", by_language), ("scenario", by_scenario)):
        for name in sorted(groups):
            print(f"{title}={name}: " + " | ".join(
                f"{provider} " + " ".join(f"{field}={groups[name][provider][field]}" for field in fields[:-1])
                for provider in providers))
    if len(providers) == 2:
        print("paired: " + ", ".join(f"jev={left}/laya={right}:{count}" for (left, right), count in sorted(paired.items())))
    print("Latency includes network and possible cache hits; this is a product-path comparison, not model-only speed.")


if __name__ == "__main__":
    raise SystemExit(main())
