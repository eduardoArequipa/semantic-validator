# Semantic Validator Python SDK

[Español](https://github.com/eduardoArequipa/semantic-validator/blob/main/sdk/python/README.es.md)
· [Source](https://github.com/eduardoArequipa/semantic-validator)
· [Runnable example](https://github.com/eduardoArequipa/semantic-validator/tree/main/examples/python-direct)

A synchronous Python SDK for semantic field validation and custom yes/no
questions using Jev. Call Jev directly with your own key, or connect to your
own Semantic Validator REST API. Requires Python 3.10+ and has no runtime
dependencies. Free and open source under Apache-2.0; Jev usage belongs to
your own TypeSafe account.

## Installation

The PyPI distribution name is `semantic-validator`; the Python import is
`semantic_validator`. The registry installation command is:

```bash
python -m pip install semantic-validator
```

For installation from a repository checkout:

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -e sdk/python
```

## Direct Jev mode

Make your own Jev key available as `TYPESAFE_API_KEY` in your backend's
environment, then run:

```python
import os
from semantic_validator import DirectValidator, SemanticValidatorError

validator = DirectValidator(jev_api_key=os.environ["TYPESAFE_API_KEY"])

try:
    result = validator.check(
        "My order arrived damaged and I want a replacement",
        "Is the customer reporting a problem with their order?",
    )
    if result.status == "uncertain":
        print("Needs review", result.confidence)
    else:
        print(result.valid, result.confidence, result.status)
except SemanticValidatorError as error:
    print("Provider request failed:", error.code)
```

Calls send your text directly to Jev and may consume credits. Keep keys in
trusted backend code. A `.env` file is not loaded automatically.

## Built-in field rules

```python
name = validator.name("Jorge Eduardo")
description = validator.validate(
    "product_description", "Cordless 18 V drill with two batteries"
)
address = validator.validate("address", "Av. Libertad 123, floor 2")
```

The built-in rules are `person_name`, `product_description`, and `address`
(v1). Read the [rule definitions and limits](https://github.com/eduardoArequipa/semantic-validator/blob/main/docs/RULES.md).
Semantic plausibility does not verify identity, product facts, or whether an
address is deliverable.

## Results and errors

Each result has `valid`, `confidence`, and `status`. A confidence score below
0.80 returns `valid=None` and `status="uncertain"`; treat it as a request for
review. Confidence is the provider's score, not a guarantee of correctness.
`SemanticValidatorError` represents request, HTTP, or response failures and
must be handled separately from an invalid value.

## Batch

```python
results = validator.validate_batch([
    {"id": "name", "rule": "person_name", "value": "Jorge Eduardo"},
    {"id": "other", "rule": "person_name", "value": "sdgfxcdg"},
])
```

Direct mode makes one Jev request per item and retains input order. Items can
have their own errors.

## Self-hosted API mode

```python
import os
from semantic_validator import Validator

validator = Validator(
    api_key=os.environ["SEMANTIC_VALIDATOR_API_KEY"],
    base_url="http://localhost:8080",
)
result = validator.name("Jorge Eduardo")
print(result.valid, result.confidence, result.status)
```

This client's API key belongs to your Semantic Validator instance. Follow
the [self-hosted guide](https://github.com/eduardoArequipa/semantic-validator/blob/main/docs/GETTING_STARTED.md).

## Development

From a checkout with the SDK installed:

```bash
python -m unittest discover -s sdk/python/tests -p 'test_*.py'
python -m unittest discover -s examples/python-direct -p 'test_*.py'
```

These tests use simulated provider responses and do not need a Jev key.
Contributions are welcome; see the [contribution guide](https://github.com/eduardoArequipa/semantic-validator/blob/main/CONTRIBUTING.md)
and [report issues](https://github.com/eduardoArequipa/semantic-validator/issues).
