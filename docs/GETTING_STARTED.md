# Getting started

[Español](../README.es.md) · [Project README](../README.md)

This guide covers the **optional self-hosted API** path. For direct SDK usage
without a server, see the [direct guide](https://validator.trialsur.cloud/docs/byok-en.html).
Obtain your own Jev key from
TypeSafe and run Semantic Validator on your machine or infrastructure. Jev's
key belongs only on the server. Your SDK uses a *different* key that you
configure for this instance. None of the SDKs is in a public package registry
yet. The legacy `Validator` classes below call your server; the new direct
clients call Jev with your key. Read the
[provider-terms caution](../README.md#bring-your-own-jev-key)
before offering an externally accessible service.

## Start your own server

Requires Go 1.22+ and a Jev account. In the repository root:

```bash
cp .env.example .env
# Edit .env: set TYPESAFE_API_KEY to YOUR Jev key.
# SEMANTIC_VALIDATOR_API_KEY=local-dev is for loopback testing only.
set -a
source .env
set +a
go run ./cmd/server
```

In another terminal, use `http://localhost:8080`. `go run` listens on all
interfaces by default; keep the port behind a local firewall while using the
example key. Alternatively, `docker compose up --build -d` publishes port 8080
on `127.0.0.1` only. Change `local-dev` before exposing the server. Never use
your Jev key as an SDK key.

For local tests, set the separate Semantic Validator key:

```bash
export SEMANTIC_VALIDATOR_API_KEY=local-dev
```

## REST API

Use one of the three registered field rules (`person_name`,
`product_description`, or `address`):

```bash
curl --fail-with-body http://localhost:8080/v1/validate \
  -H "Authorization: Bearer $SEMANTIC_VALIDATOR_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"rule":"person_name","value":"Jorge Eduardo"}'
```

Ask a custom yes/no question for a form, chatbot, or workflow:

```bash
curl --fail-with-body http://localhost:8080/v1/check \
  -H "Authorization: Bearer $SEMANTIC_VALIDATOR_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"value":"A cordless 18 V drill with two batteries","question":"Does the text identify a product and at least one feature?"}'
```

The `POST /v1/validate/batch` endpoint handles 1–100 rule checks. It does not
batch unrelated custom questions. The full contract is in
[OpenAPI](../internal/api/openapi.yaml).

## TypeScript / Node.js

Build the SDK from [`sdk/typescript`](../sdk/typescript/README.md) and install
it in your Node.js project. The SDK requires Node.js 18 or newer.

```bash
cd sdk/typescript
npm install
npm run build
# In your application directory, install the local package path:
npm install /path/to/semantic-validator/sdk/typescript
```

```ts
import { Validator } from "@semantic-validator/sdk";

const validator = new Validator({
  apiKey: process.env.SEMANTIC_VALIDATOR_API_KEY!,
  baseURL: "http://localhost:8080",
});

const result = await validator.check(
  "A cordless 18 V drill with two batteries",
  "Does the text identify a product and at least one feature?",
);
console.log(result.valid, result.confidence, result.status);
```

Run this in a configured Node.js TypeScript project with Node type definitions;
the `!` annotation does not load an environment variable. If you need a complete
starter project, see the [TypeScript demo](../examples/typescript-demo/README.md).

## Python

Python 3.10 or newer. From the repository root:

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install ./sdk/python
```

```python
import os
from semantic_validator import Validator

validator = Validator(
    api_key=os.environ["SEMANTIC_VALIDATOR_API_KEY"],
    base_url="http://localhost:8080",
)
result = validator.check(
    "A cordless 18 V drill with two batteries",
    "Does the text identify a product and at least one feature?",
)
print(result.valid, result.confidence, result.status)
```

## Go

Go 1.22 or newer. Add the public Go SDK to your application:

```bash
go mod init example-validator
go get github.com/eduardoArequipa/semantic-validator/sdk/go@latest
```

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    validator "github.com/eduardoArequipa/semantic-validator/sdk/go"
)

func main() {
    client, err := validator.NewClientWithOptions(
        os.Getenv("SEMANTIC_VALIDATOR_API_KEY"),
        validator.Options{BaseURL: "http://localhost:8080"},
    )
    if err != nil { log.Fatal(err) }
    result, err := client.Check(context.Background(),
        "A cordless 18 V drill with two batteries",
        "Does the text identify a product and at least one feature?")
    if err != nil { log.Fatal(err) }
    if result.Valid == nil {
        fmt.Println("Needs review", result.Confidence)
    } else {
        fmt.Println(*result.Valid, result.Confidence)
    }
}
```

Save the example as `main.go`, then run `go mod tidy` and `go run .`.

## Java

Java 11 or newer and Maven. From the repository root, install the SDK into
your local Maven repository:

```bash
mvn -f sdk/java/pom.xml install
```

In your application's `pom.xml`:

```xml
<dependency>
  <groupId>io.semanticvalidator</groupId>
  <artifactId>semantic-validator-sdk</artifactId>
  <version>0.3.0</version>
</dependency>
```

```java
import io.semanticvalidator.sdk.SemanticValidatorClient;
import java.time.Duration;

var client = new SemanticValidatorClient(
    System.getenv("SEMANTIC_VALIDATOR_API_KEY"),
    "http://localhost:8080",
    Duration.ofSeconds(40)
);
var result = client.check("A cordless 18 V drill with two batteries",
    "Does the text identify a product and at least one feature?");
System.out.println(result.getValid()); // May be null when uncertain.
```

Put this snippet inside a method in your Maven application. See the
[Java SDK README](../sdk/java/README.md) for the complete class structure.

## Results, quotas, and boundaries

| `status` | `valid` | Meaning |
| --- | --- | --- |
| `valid` | `true` | The provider answered yes. |
| `invalid` | `false` | The provider answered no; this is not an API error. |
| `uncertain` | `null` | Request more context or human review. |

`confidence` is the provider score, not a guaranteed probability of being
right. A plausible name does not establish someone's identity. Validate empty
fields, length, and format with ordinary code first. Handle HTTP errors and
uncertain outcomes separately.

If you configure `API_KEYS_FILE`, your instance can issue individual keys with
per-key daily quotas. The example single-key setup does not set a daily quota.
The API limits authenticated requests to 60 per minute per IP by default.

Store keys only on your backend. Text sent to your server may be processed by
Jev. Use fictional text in sample requests. When finished, run
`unset SEMANTIC_VALIDATOR_API_KEY`.
