import contextlib
import io
import json
import os
import sys
import tempfile
import types
import unittest
from pathlib import Path
from unittest.mock import patch

import evaluate


class FakeResponse:
    def __init__(self, payload):
        self.payload = payload

    def __enter__(self):
        return self

    def __exit__(self, *_args):
        return False

    def read(self):
        return json.dumps(self.payload).encode()


class EvaluationTests(unittest.TestCase):
    def test_examples_and_dry_run_do_not_need_key(self):
        cases = evaluate.load_cases("evals/cases.example.json")
        self.assertEqual(len(cases), 10)
        output = io.StringIO()
        with patch.dict(os.environ, {}, clear=True), contextlib.redirect_stdout(output):
            result = evaluate.main(["--dry-run"])
        self.assertEqual(result, 0)
        self.assertIn("no API requests", output.getvalue())

    def test_field_rule_dataset_has_balanced_labels_and_dry_run(self):
        cases = evaluate.load_cases("evals/field-rules.json")
        self.assertEqual(len(cases), 90)
        for rule in ("person_name", "product_description", "address"):
            labels = [case["expected_valid"] for case in cases if case["rule"] == rule]
            self.assertEqual(len(labels), 30)
            self.assertEqual(sum(label is True for label in labels), 12)
            self.assertEqual(sum(label is False for label in labels), 12)
            self.assertEqual(sum(label is None for label in labels), 6)
        output = io.StringIO()
        with patch.dict(os.environ, {}, clear=True), contextlib.redirect_stdout(output):
            result = evaluate.main(["--cases", "evals/field-rules.json", "--dry-run"])
        self.assertEqual(result, 0)
        self.assertIn("Ready: 90 cases", output.getvalue())

    def test_small_sample_covers_each_rule_and_label(self):
        cases = evaluate.load_cases("evals/field-rules.json")
        selected = evaluate.select_cases(cases, sample_per_label=1)
        self.assertEqual(len(selected), 9)
        self.assertEqual(len({case["id"] for case in selected}), 9)
        for rule in ("person_name", "product_description", "address"):
            self.assertEqual({case["expected_valid"] for case in selected if case["rule"] == rule},
                             {True, False, None})
        self.assertEqual({case["rule"] for case in selected[:3]},
                         {"person_name", "product_description", "address"})
        self.assertEqual({case["expected_valid"] for case in selected[:3]},
                         {True, False, None})
        output = io.StringIO()
        with patch.dict(os.environ, {}, clear=True), contextlib.redirect_stdout(output):
            result = evaluate.main(["--cases", "evals/field-rules.json", "--sample-per-label", "1", "--dry-run"])
        self.assertEqual(result, 0)
        self.assertIn("Ready: 9 cases", output.getvalue())
        self.assertIn("Selected IDs:", output.getvalue())
        with self.assertRaises(ValueError):
            evaluate.select_cases(cases, sample_per_label=7)

    def test_ambiguous_cases_are_reported_separately_by_rule(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "cases.json"
            path.write_text(json.dumps([
                {"id": "ambiguous", "value": "Sol", "rule": "person_name", "expected_valid": None},
            ]), encoding="utf-8")
            output = io.StringIO()
            with patch.dict(os.environ, {"SEMANTIC_VALIDATOR_API_KEY": "test-key"}), \
                 patch.object(evaluate, "open_request", return_value=FakeResponse(
                     {"valid": None, "status": "uncertain", "confidence": 0.63})), \
                 contextlib.redirect_stdout(output):
                result = evaluate.main(["--cases", str(path)])
            self.assertEqual(result, 0)
            self.assertIn("ambiguous_uncertain=1", output.getvalue())
            self.assertIn("Rule person_name:", output.getvalue())

    def test_direct_mode_requires_own_key_and_uses_client_contract(self):
        output = io.StringIO()
        with patch.dict(os.environ, {}, clear=True), contextlib.redirect_stderr(output):
            self.assertEqual(evaluate.main(["--direct", "--cases", "evals/field-rules.json"]), 2)
        self.assertIn("TYPESAFE_API_KEY", output.getvalue())

        class FakeClient:
            def validate(self, rule, value):
                self.rule = rule
                self.value = value
                return type("Result", (), {"valid": True, "status": "valid", "confidence": 0.94})()

        client = FakeClient()
        self.assertEqual(evaluate.evaluate_direct_case(client, {
            "rule": "address", "value": "Av. Libertad 123"}), (True, "valid", 0.94))
        self.assertEqual(client.rule, "address")

    def test_direct_mode_stops_after_rate_limit(self):
        calls = []

        class RateLimited(Exception):
            status_code = 429

        class FakeDirectClient:
            def __init__(self, **kwargs):
                pass

            def validate(self, rule, value):
                calls.append(rule)
                raise RateLimited()

        fake_sdk = types.ModuleType("semantic_validator")
        fake_sdk.DirectValidator = FakeDirectClient
        fake_sdk.SemanticValidatorError = RateLimited
        output = io.StringIO()
        with patch.dict(sys.modules, {"semantic_validator": fake_sdk}), \
             patch.dict(os.environ, {"TYPESAFE_API_KEY": "test-key"}), \
             contextlib.redirect_stdout(output):
            result = evaluate.main(["--direct", "--cases", "evals/field-rules.json",
                                    "--sample-per-label", "1"])
        self.assertEqual(result, 1)
        self.assertEqual(len(calls), 1)
        self.assertIn("stopping", output.getvalue())

    def test_validation_rejects_duplicate_ids_and_exposed_text_in_ids(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "cases.json"
            case = {"id": "safe", "value": "Texto ficticio", "question": "¿Es claro?", "expected_valid": True}
            path.write_text(json.dumps([case, case]), encoding="utf-8")
            with self.assertRaises(ValueError):
                evaluate.load_cases(path)
            case["id"] = "private text with spaces"
            path.write_text(json.dumps([case]), encoding="utf-8")
            with self.assertRaises(ValueError):
                evaluate.load_cases(path)

    def test_summary_distinguishes_errors_and_uncertain_without_text_or_key(self):
        secret = "test-secret-token"
        value = "Mensaje ficticio privado"
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "cases.json"
            path.write_text(json.dumps([
                {"id": "positivo", "value": value, "question": "¿Está claro?", "expected_valid": True},
                {"id": "incierto", "value": value, "question": "¿Está claro?", "expected_valid": False},
            ]), encoding="utf-8")
            captured = []

            def respond(request, timeout):
                captured.append(request)
                return FakeResponse({"valid": True, "status": "valid", "confidence": 0.91} if len(captured) == 1 else
                                    {"valid": None, "status": "uncertain", "confidence": 0.62})

            output = io.StringIO()
            with patch.dict(os.environ, {"SEMANTIC_VALIDATOR_API_KEY": secret}), \
                 patch.object(evaluate, "open_request", respond), contextlib.redirect_stdout(output):
                result = evaluate.main(["--cases", str(path)])
            self.assertEqual(result, 0)
            self.assertEqual(len(captured), 2)
            self.assertEqual(captured[0].headers["Authorization"], "Bearer " + secret)
            self.assertIn("correct=1", output.getvalue())
            self.assertIn("uncertain=1", output.getvalue())
            self.assertNotIn(secret, output.getvalue())
            self.assertNotIn(value, output.getvalue())

    def test_invalid_api_response_is_error_not_wrong_answer(self):
        case = {"value": "Ejemplo", "question": "¿Es claro?"}
        with patch.object(evaluate, "open_request", return_value=FakeResponse(
                {"valid": True, "status": "invalid", "confidence": 1})):
            with self.assertRaisesRegex(ValueError, "invalid API response"):
                evaluate.evaluate_case("https://validator.trialsur.cloud", "test", case, 1)


if __name__ == "__main__":
    unittest.main()
