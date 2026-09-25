import contextlib
import io
import json
import os
import tempfile
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
