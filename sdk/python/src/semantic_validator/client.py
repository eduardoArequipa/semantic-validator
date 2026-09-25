from __future__ import annotations

import json
from concurrent.futures import ThreadPoolExecutor
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Any, Callable, Iterable, Mapping, Optional
from urllib.parse import urlsplit


class SemanticValidatorError(Exception):
    """Error returned by Semantic Validator or raised by the SDK transport."""

    def __init__(self, message: str, *, code: str = "client_error", status_code: Optional[int] = None):
        super().__init__(message)
        self.code = code
        self.status_code = status_code


@dataclass(frozen=True)
class ValidationResult:
    valid: Optional[bool]
    confidence: float
    status: str

    @classmethod
    def from_payload(cls, payload: Mapping[str, Any]) -> "ValidationResult":
        try:
            return cls(
                valid=payload["valid"],
                confidence=float(payload["confidence"]),
                status=str(payload["status"]),
            )
        except (KeyError, TypeError, ValueError) as exc:
            raise SemanticValidatorError("invalid validation response", code="invalid_response") from exc


@dataclass(frozen=True)
class BatchItem:
    id: str
    rule: str
    value: str

    def to_payload(self) -> dict[str, str]:
        return {"id": self.id, "rule": self.rule, "value": self.value}


@dataclass(frozen=True)
class BatchResult:
    id: str
    valid: Optional[bool]
    confidence: float
    status: str
    error: Optional[dict[str, str]] = None

    @classmethod
    def from_payload(cls, payload: Mapping[str, Any]) -> "BatchResult":
        try:
            error = payload.get("error")
            return cls(
                id=str(payload["id"]),
                valid=payload.get("valid"),
                confidence=float(payload.get("confidence", 0)),
                status=str(payload["status"]),
                error=error if isinstance(error, dict) else None,
            )
        except (KeyError, TypeError, ValueError) as exc:
            raise SemanticValidatorError("invalid batch response", code="invalid_response") from exc


Opener = Callable[..., Any]


class _NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, request, fp, code, message, headers, newurl):
        return None


class Validator:
    """Synchronous client for the Semantic Validator API."""

    def __init__(
        self,
        api_key: str,
        *,
        base_url: str = "http://localhost:8080",
        timeout: float = 30.0,
        opener: Optional[Opener] = None,
    ):
        if not api_key:
            raise ValueError("api_key is required")
        if timeout <= 0:
            raise ValueError("timeout must be positive")
        self._api_key = api_key
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout
        self._opener = opener or urllib.request.urlopen

    def validate(self, rule: str, value: str) -> ValidationResult:
        payload = self._post("/v1/validate", {"rule": rule, "value": value})
        return ValidationResult.from_payload(payload)

    def name(self, value: str) -> ValidationResult:
        return self.validate("person_name", value)

    def check(self, value: str, question: str) -> ValidationResult:
        payload = self._post("/v1/check", {"value": value, "question": question})
        return ValidationResult.from_payload(payload)

    def validate_batch(self, items: Iterable[BatchItem | Mapping[str, str]]) -> list[BatchResult]:
        normalized = [self._normalize_batch_item(item) for item in items]
        if not 1 <= len(normalized) <= 100:
            raise ValueError("items must contain between 1 and 100 elements")
        payload = self._post("/v1/validate/batch", {"items": [item.to_payload() for item in normalized]})
        response_items = payload.get("items")
        if not isinstance(response_items, list):
            raise SemanticValidatorError("invalid batch response", code="invalid_response")
        return [BatchResult.from_payload(item) for item in response_items]

    def _normalize_batch_item(self, item: BatchItem | Mapping[str, str]) -> BatchItem:
        if isinstance(item, BatchItem):
            return item
        if isinstance(item, Mapping):
            try:
                return BatchItem(id=str(item["id"]), rule=str(item["rule"]), value=str(item["value"]))
            except KeyError as exc:
                raise ValueError(f"batch item requires {exc.args[0]}") from exc
        raise TypeError("batch items must be BatchItem instances or mappings")

    def _post(self, path: str, payload: Mapping[str, Any]) -> dict[str, Any]:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        request = urllib.request.Request(
            self._base_url + path,
            data=body,
            headers={
                "Accept": "application/json",
                "Authorization": f"Bearer {self._api_key}",
                "Content-Type": "application/json",
            },
            method="POST",
        )
        try:
            response = self._opener(request, timeout=self._timeout)
            raw_body = response.read()
        except urllib.error.HTTPError as exc:
            raw_body = exc.read()
            raise self._error_from_body(raw_body, exc.code) from exc
        except urllib.error.URLError as exc:
            raise SemanticValidatorError(f"request failed: {exc.reason}", code="transport_error") from exc
        except TimeoutError as exc:
            raise SemanticValidatorError("request timed out", code="timeout") from exc
        finally:
            if "response" in locals() and hasattr(response, "close"):
                response.close()

        try:
            decoded = json.loads(raw_body.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError) as exc:
            raise SemanticValidatorError("server returned invalid JSON", code="invalid_response") from exc
        if not isinstance(decoded, dict):
            raise SemanticValidatorError("server returned an invalid response", code="invalid_response")
        return decoded

    @staticmethod
    def _error_from_body(body: bytes, status_code: int) -> SemanticValidatorError:
        try:
            decoded = json.loads(body.decode("utf-8"))
            error = decoded.get("error", {})
            code = str(error.get("code", "http_error"))
            message = str(error.get("message", "request failed"))
        except (UnicodeDecodeError, json.JSONDecodeError, AttributeError):
            code = "http_error"
            message = "request failed"
        return SemanticValidatorError(message, code=code, status_code=status_code)


class DirectValidator:
    """Synchronous, server-side client that sends your own Jev key straight to Jev."""

    _name_question = "¿Este texto parece representar el nombre de una persona?"

    def __init__(self, jev_api_key: str, *, base_url: str = "https://api.typesafe.ai",
                 timeout: float = 30.0, opener: Optional[Opener] = None):
        if not isinstance(jev_api_key, str) or not jev_api_key.strip():
            raise ValueError("jev_api_key is required")
        if timeout <= 0:
            raise ValueError("timeout must be positive")
        parsed = urlsplit(base_url)
        if not ((parsed.scheme == "https" and parsed.hostname == "api.typesafe.ai") or
                (parsed.scheme == "http" and parsed.hostname in {"localhost", "127.0.0.1", "::1"})):
            raise ValueError("Jev endpoint must be api.typesafe.ai (except loopback tests)")
        if not parsed.hostname or parsed.username or parsed.password or parsed.query or parsed.fragment or parsed.path not in ("", "/"):
            raise ValueError("invalid Jev endpoint")
        self._api_key = jev_api_key.strip()
        self._base_url = base_url.rstrip("/")
        self._timeout = timeout
        self._opener = opener or urllib.request.build_opener(_NoRedirect()).open

    def validate(self, rule: str, value: str) -> ValidationResult:
        self._valid_text(value, "value")
        if rule != "person_name":
            raise SemanticValidatorError("requested rule is not registered", code="unknown_rule")
        return self.check(value, self._name_question)

    def name(self, value: str) -> ValidationResult:
        return self.validate("person_name", value)

    def check(self, value: str, question: str) -> ValidationResult:
        state = self._valid_text(value, "value")
        instructions = self._valid_text(question, "question")
        body = json.dumps({"model": "jev-latest", "state": state, "questions": {"result": {
            "type": "choice", "instructions": instructions,
            "criteria": {"true": "El texto cumple la condición.", "false": "El texto no cumple la condición."},
        }}}, ensure_ascii=False).encode("utf-8")
        request = urllib.request.Request(self._base_url + "/v1/systemone", data=body,
            headers={"Accept": "application/json", "Authorization": f"Bearer {self._api_key}",
                     "Content-Type": "application/json"}, method="POST")
        try:
            response = self._opener(request, timeout=self._timeout)
            try:
                raw = response.read(2 * 1024 * 1024 + 1)
            finally:
                response.close()
        except urllib.error.HTTPError as exc:
            raise SemanticValidatorError(f"Jev returned HTTP {exc.code}", code="provider_error", status_code=exc.code) from exc
        except TimeoutError as exc:
            raise SemanticValidatorError("Jev request timed out", code="provider_timeout") from exc
        except urllib.error.URLError as exc:
            code = "provider_timeout" if isinstance(exc.reason, TimeoutError) else "provider_error"
            raise SemanticValidatorError("Jev request failed", code=code) from exc
        if len(raw) > 2 * 1024 * 1024:
            raise SemanticValidatorError("Jev response too large", code="invalid_response")
        try:
            payload = json.loads(raw)
            answer = payload["answers"]["result"]
            choice = answer["choice"]
            confidence = answer.get("confidence", answer.get("probabilities", {}).get(choice))
        except (ValueError, TypeError, KeyError, AttributeError) as exc:
            raise SemanticValidatorError("invalid Jev response", code="invalid_response") from exc
        if choice not in ("true", "false") or isinstance(confidence, bool) or not isinstance(confidence, (int, float)) or not 0 <= confidence <= 1:
            raise SemanticValidatorError("invalid Jev response", code="invalid_response")
        confidence = float(confidence)
        if confidence < 0.8:
            return ValidationResult(valid=None, confidence=confidence, status="uncertain")
        valid = choice == "true"
        return ValidationResult(valid=valid, confidence=confidence, status="valid" if valid else "invalid")

    def validate_batch(self, items: Iterable[BatchItem | Mapping[str, str]]) -> list[BatchResult]:
        normalized = [Validator._normalize_batch_item(self, item) for item in items]
        if not 1 <= len(normalized) <= 100:
            raise ValueError("items must contain between 1 and 100 elements")
        ids = set()
        for item in normalized:
            if not item.id or not isinstance(item.rule, str) or not isinstance(item.value, str) or item.id in ids:
                raise ValueError("batch items require unique ids, a rule, and a string value")
            ids.add(item.id)

        def evaluate(item: BatchItem) -> BatchResult:
            try:
                result = self.validate(item.rule, item.value)
                return BatchResult(id=item.id, valid=result.valid, confidence=result.confidence, status=result.status)
            except SemanticValidatorError as exc:
                message = "requested rule is not registered" if exc.code == "unknown_rule" else "invalid value" if exc.code == "invalid_request" else "semantic provider request failed"
                return BatchResult(id=item.id, valid=None, confidence=0, status="error", error={"code": exc.code, "message": message})

        with ThreadPoolExecutor(max_workers=min(4, len(normalized))) as pool:
            return list(pool.map(evaluate, normalized))

    @staticmethod
    def _valid_text(value: str, field: str) -> str:
        if not isinstance(value, str) or not value.strip() or len(value.strip()) > 10_000:
            raise SemanticValidatorError(f"{field} must contain between 1 and 10000 characters", code="invalid_request")
        return value.strip()
