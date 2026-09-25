# Semantic Validator SDK for Java

Maven clients for direct Jev BYOK and the optional self-hosted Semantic Validator API. Requires Java 11+.

## Direct Jev mode (recommended)

```java
import io.semanticvalidator.sdk.DirectValidator;

var validator = new DirectValidator(System.getenv("TYPESAFE_API_KEY"));
var result = validator.check("My order arrived damaged", "Is this a complaint?");
System.out.println(result.getValid());
```

Run on a trusted backend. `name`, `validate`, and `validateBatch` are available; each batch item calls Jev separately. Confidence below 0.80 gives `getValid() == null`.

## Build locally

```bash
cd sdk/java
mvn package
```

## Own server example (optional)

```java
import io.semanticvalidator.sdk.SemanticValidatorClient;
import io.semanticvalidator.sdk.ValidationResult;

import java.io.IOException;

public class Example {
    public static void main(String[] args) throws IOException, InterruptedException {
        String apiKey = System.getenv("SEMANTIC_VALIDATOR_API_KEY");
        SemanticValidatorClient validator = new SemanticValidatorClient(apiKey);

        ValidationResult result = validator.name("Jorge Eduardo");
        System.out.println(result.getValid());
        System.out.println(result.getConfidence());
        System.out.println(result.getStatus());
    }
}
```

Use `validate(rule, value)` for registered rules, `check(value, question)` for custom semantic questions, and `validateBatch(items)` for batches of up to 100 items. `ValidationResult.getValid()` returns `null` when the result is uncertain. Batch item failures are included in `BatchResult.getError()`; HTTP failures throw `SemanticValidatorException` with an error code and status code.

The default endpoint is `http://localhost:8080` with a 30-second timeout. Use `new SemanticValidatorClient(apiKey, baseUrl, timeout)` to customize them. Keep the API key on a trusted server; do not ship it in a desktop or mobile application.
