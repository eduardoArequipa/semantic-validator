import contextlib
import io
import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import compare_providers


class FakeResponse:
    def __init__(self, payload):
        self.payload = payload

    def __enter__(self):
        return self

    def __exit__(self, *_args):
        return False

    def read(self):
        return json.dumps(self.payload).encode()


class ComparisonTests(unittest.TestCase):
    def test_dataset_has_100_balanced_cases(self):
        cases = compare_providers.load_cases("evals/provider-comparison.example.json")
        self.assertEqual(len(cases), 100)
        self.assertEqual(len({case["id"] for case in cases}), 100)
        for language in ("es", "en"):
            selected = [case for case in cases if case["language"] == language]
            self.assertEqual(len(selected), 50)
            self.assertEqual(sum(case["expected_valid"] for case in selected), 25)
        for scenario in ("person_name", "complaint", "purchase_intent", "product_description", "failure_report"):
            self.assertEqual(sum(case["scenario"] == scenario for case in cases), 20)

    def test_dry_run_needs_no_key(self):
        output = io.StringIO()
        with patch.dict(os.environ, {}, clear=True), contextlib.redirect_stdout(output):
            result = compare_providers.main(["--dry-run"])
        self.assertEqual(result, 0)
        self.assertIn("100 labeled cases; 200 requests maximum; none sent", output.getvalue())

    def test_laya_request_uses_noul_and_no_jev_key(self):
        case = {"value": "I want to buy one", "question": "Does the customer want to buy?", "language": "en"}
        captured = []

        def respond(request, timeout):
            captured.append(request)
            return FakeResponse({"answers": {"result": {"noul": 0.92}}})

        with patch.object(compare_providers.evaluate, "open_request", respond):
            result = compare_providers.evaluate_laya("http://127.0.0.1:8000", "", case, 1)
        self.assertEqual(result, (True, "valid", 0.92))
        payload = json.loads(captured[0].data)
        self.assertEqual(payload["questions"]["result"]["type"], "noul")
        self.assertNotIn("Authorization", captured[0].headers)
        self.assertEqual(captured[0].full_url, "http://127.0.0.1:8000/v1/systemone")

    def test_uncertain_and_false_probabilities(self):
        self.assertEqual(compare_providers.parse_laya_response({"answers": {"result": {"noul": 0.19}}}),
                         (False, "invalid", 0.81))
        self.assertEqual(compare_providers.parse_laya_response({"answers": {"result": {"noul": 0.55}}}),
                         (None, "uncertain", 0.55))
        for probability in (None, "0.9", float("nan"), 1.1):
            with self.assertRaises(ValueError):
                compare_providers.parse_laya_response({"answers": {"result": {"noul": probability}}})

    def test_choice_uses_answer_probability_not_entropy_confidence(self):
        response = {"answers": {"result": {"choice": "B", "confidence": 0.02,
                                           "answer_confidence": 0.93}}}
        self.assertEqual(compare_providers.parse_laya_response(response, "choice"), (False, "invalid", 0.93))
        response["answers"]["result"]["choice"] = "wrong"
        with self.assertRaises(ValueError):
            compare_providers.parse_laya_response(response, "choice")

    def test_choice_request_has_neutral_options(self):
        case = {"value": "Quiero comprar uno", "question": "¿Desea comprar?", "language": "es"}
        captured = []

        def respond(request, timeout):
            captured.append(request)
            return FakeResponse({"answers": {"result": {"choice": "A", "answer_confidence": 0.9}}})

        with patch.object(compare_providers.evaluate, "open_request", respond):
            result = compare_providers.evaluate_laya("http://127.0.0.1:8000", "", case, 1, "choice")
        self.assertEqual(result, (True, "valid", 0.9))
        question = json.loads(captured[0].data)["questions"]["result"]
        self.assertEqual(question["type"], "choice")
        self.assertEqual(set(question["criteria"]), {"A", "B"})

    def test_laya_only_dry_run_counts_one_request_per_case(self):
        output = io.StringIO()
        with patch.dict(os.environ, {}, clear=True), contextlib.redirect_stdout(output):
            result = compare_providers.main(["--dry-run", "--laya-only", "--laya-primitive", "choice"])
        self.assertEqual(result, 0)
        self.assertIn("100 requests maximum", output.getvalue())

    def test_rejects_bad_dataset_and_nonlocal_http(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "cases.json"
            path.write_text(json.dumps([{"scenario": "unsafe text", "language": "es", "rule": "person_name",
                                         "examples": [["Ana López", True]]}]), encoding="utf-8")
            with self.assertRaises(ValueError):
                compare_providers.load_cases(path)
        output = io.StringIO()
        with contextlib.redirect_stderr(output):
            self.assertEqual(compare_providers.main(["--dry-run", "--laya-base-url", "http://example.com"]), 2)
        self.assertIn("HTTP is allowed only for local development", output.getvalue())

    def test_summary_does_not_print_text_or_keys(self):
        value = "Mensaje privado ficticio"
        data = [{"scenario": "complaint", "language": "es", "question": "¿Hay reclamo?",
                 "examples": [[value, True]]}]
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "cases.json"
            path.write_text(json.dumps(data), encoding="utf-8")
            output = io.StringIO()
            with patch.dict(os.environ, {"SEMANTIC_VALIDATOR_API_KEY": "secret-key"}), \
                 patch.object(compare_providers.evaluate, "evaluate_case", return_value=(True, "valid", 0.9)), \
                 patch.object(compare_providers, "evaluate_laya", return_value=(False, "invalid", 0.95)), \
                 contextlib.redirect_stdout(output):
                result = compare_providers.main(["--cases", str(path)])
        self.assertEqual(result, 0)
        self.assertIn("jev: correct=1", output.getvalue())
        self.assertIn("laya: correct=0, false_positive=0, false_negative=1", output.getvalue())
        self.assertNotIn(value, output.getvalue())
        self.assertNotIn("secret-key", output.getvalue())


if __name__ == "__main__":
    unittest.main()
