import json
import unittest

from semantic_validator import BatchItem, DirectValidator, SemanticValidatorError, Validator


class FakeResponse:
    def __init__(self, payload):
        self._body = json.dumps(payload).encode("utf-8")

    def read(self, size=-1):
        return self._body

    def close(self):
        pass


class FakeTransport:
    def __init__(self):
        self.requests = []

    def __call__(self, request, timeout):
        self.requests.append((request, timeout))
        if request.full_url.endswith("/v1/validate"):
            return FakeResponse({"valid": True, "confidence": 0.98, "status": "valid"})
        if request.full_url.endswith("/v1/check"):
            return FakeResponse({"valid": None, "confidence": 0.63, "status": "uncertain"})
        return FakeResponse({"items": [{"id": "name", "valid": True, "confidence": 1, "status": "valid"}]})


class ValidatorTests(unittest.TestCase):
    def setUp(self):
        self.transport = FakeTransport()
        self.client = Validator("test-key", base_url="http://test", opener=self.transport)

    def test_name(self):
        result = self.client.name("Jorge Eduardo")
        self.assertTrue(result.valid)
        self.assertEqual(result.confidence, 0.98)
        self.assertEqual(self.transport.requests[0][0].get_method(), "POST")

    def test_check_uncertain(self):
        result = self.client.check("texto", "¿La condición se cumple?")
        self.assertIsNone(result.valid)
        self.assertEqual(result.status, "uncertain")

    def test_batch(self):
        results = self.client.validate_batch([BatchItem("name", "person_name", "Jorge Eduardo")])
        self.assertEqual(results[0].id, "name")

    def test_requires_api_key(self):
        with self.assertRaises(ValueError):
            Validator("")

    def test_empty_batch(self):
        with self.assertRaises(ValueError):
            self.client.validate_batch([])


class DirectValidatorTests(unittest.TestCase):
    def test_direct_request_and_status(self):
        calls = []

        def opener(request, timeout):
            calls.append(request)
            return FakeResponse({"answers": {"result": {"type": "choice", "choice": "true", "confidence": 0.96}}})

        client = DirectValidator("own-key", opener=opener)
        result = client.name(" Jorge Eduardo ")
        self.assertTrue(result.valid)
        self.assertEqual(result.status, "valid")
        self.assertEqual(calls[0].full_url, "https://api.typesafe.ai/v1/systemone")
        self.assertEqual(calls[0].get_header("Authorization"), "Bearer own-key")
        body = json.loads(calls[0].data)
        self.assertEqual(body["state"], "Jorge Eduardo")
        self.assertEqual(body["model"], "jev-latest")
        self.assertEqual(body["questions"]["result"]["type"], "choice")

    def test_invalid_uncertain_and_batch(self):
        def opener(request, timeout):
            state = json.loads(request.data)["state"]
            choice, confidence = ("false", 0.95) if state == "No" else ("true", 0.63)
            return FakeResponse({"answers": {"result": {"choice": choice, "confidence": confidence}}})

        client = DirectValidator("own-key", opener=opener)
        self.assertFalse(client.check("No", "¿Compra?").valid)
        self.assertIsNone(client.check("texto", "pregunta").valid)
        results = client.validate_batch([BatchItem("a", "person_name", "No"),
                                         BatchItem("b", "missing", "x"),
                                         BatchItem("c", "person_name", " ")])
        self.assertEqual([item.id for item in results], ["a", "b", "c"])
        self.assertEqual(results[1].error["code"], "unknown_rule")
        self.assertEqual(results[2].error["code"], "invalid_request")

    def test_errors_do_not_expose_key(self):
        with self.assertRaises(ValueError):
            DirectValidator("")
        with self.assertRaises(ValueError):
            DirectValidator("own-key", base_url="http://example.com")
        client = DirectValidator("own-key", opener=lambda request, timeout: FakeResponse({}))
        with self.assertRaises(SemanticValidatorError) as caught:
            client.check("text", "question")
        self.assertEqual(caught.exception.code, "invalid_response")
        self.assertNotIn("own-key", str(caught.exception))


if __name__ == "__main__":
    unittest.main()
