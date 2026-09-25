# Semantic Validator Go SDK

Go clients for direct Jev BYOK and the optional self-hosted Semantic Validator API. Standard library only.

Install with Go 1.22 or newer:

```bash
go get github.com/eduardoArequipa/semantic-validator/sdk/go@latest
```

Import as `validator "github.com/eduardoArequipa/semantic-validator/sdk/go"`.

## Direct Jev mode (recommended)

```go
client, err := validator.NewDirectClient(os.Getenv("TYPESAFE_API_KEY"))
if err != nil { log.Fatal(err) }
result, err := client.Check(context.Background(), "My order arrived damaged", "Is this a complaint?")
if err != nil { log.Fatal(err) }
fmt.Println(result.Valid, result.Confidence, result.Status)
```

Run on a trusted backend. `Name`, `Validate`, and `ValidateBatch` are available; each batch item calls Jev separately. A confidence below 0.80 gives `Valid == nil`.

## Own server mode (optional)

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
	client := validator.NewClient(os.Getenv("SEMANTIC_VALIDATOR_API_KEY"))
	result, err := client.Name(context.Background(), "Jorge Eduardo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Valid, result.Confidence, result.Status)
}
```

`NewClient` uses `http://localhost:8080` with a 30-second timeout. Use `NewClientWithOptions` to set a different base URL, timeout, or reusable `http.Client`. The methods `Validate`, `Name`, `Check`, and `ValidateBatch` accept a context. An uncertain result has `Valid == nil`; per-item batch errors are returned in `BatchResult.Error`. API-level HTTP failures can be inspected with `errors.As(err, &apiErr)` where `apiErr` is `*validator.APIError`.
