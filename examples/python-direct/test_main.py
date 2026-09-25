import io
import json
import unittest
from urllib.error import URLError

from semantic_validator import DirectValidator, SemanticValidatorError

from main import run


class ExampleTests(unittest.TestCase):
    def test_runs_without_a_real_key_or_network(self):
        requests = []

        def fake_open(request, *, timeout):
            self.assertEqual(request.full_url, "https://api.typesafe.ai/v1/systemone")
            self.assertEqual(request.get_header("Authorization"), "Bearer test-only")
            payload = json.loads(request.data)
            requests.append(payload)
            is_name = "nombre de una persona" in payload["questions"]["result"]["instructions"]
            answer = {"choice": "true", "confidence": 0.96 if is_name else 0.62}
            return io.BytesIO(json.dumps({"answers": {"result": answer}}).encode())

        output = []
        run(DirectValidator("test-only", opener=fake_open), output.append)
        self.assertEqual(len(requests), 2)
        self.assertIn("nombre: True", output[0])
        self.assertIn("reclamo: requiere revisión", output[1])

    def test_provider_failure_is_not_false(self):
        def fake_open(request, *, timeout):
            raise URLError("offline")

        with self.assertRaises(SemanticValidatorError) as raised:
            run(DirectValidator("test-only", opener=fake_open))
        self.assertEqual(raised.exception.code, "provider_error")


if __name__ == "__main__":
    unittest.main()
