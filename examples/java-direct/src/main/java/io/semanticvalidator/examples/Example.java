package io.semanticvalidator.examples;

import io.semanticvalidator.sdk.DirectValidator;
import io.semanticvalidator.sdk.ValidationResult;

import java.io.IOException;
import java.io.PrintStream;

/** Calls Jev only when run with the user's own key. */
public final class Example {
    private Example() {}

    public static void main(String[] args) throws IOException, InterruptedException {
        String key = System.getenv("TYPESAFE_API_KEY");
        if (key == null || key.isBlank()) {
            throw new IllegalStateException("TYPESAFE_API_KEY is required; run Maven tests first");
        }
        run(new DirectValidator(key), System.out);
    }

    public static void run(DirectValidator validator, PrintStream output)
            throws IOException, InterruptedException {
        show(output, "nombre", validator.name("Jorge Eduardo"));
        show(output, "reclamo", validator.check(
                "Mi pedido llegó dañado y quiero una solución",
                "¿El cliente está presentando un reclamo?"));
    }

    private static void show(PrintStream output, String label, ValidationResult result) {
        if ("uncertain".equals(result.getStatus())) {
            output.printf("%s: requiere revisión (confianza %.2f)%n", label, result.getConfidence());
        } else {
            output.printf("%s: %s (confianza %.2f)%n", label, result.getValid(), result.getConfidence());
        }
    }
}
