package io.semanticvalidator.sdk;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

/** Server-side client that uses the caller's own Jev key without our API. */
public final class DirectValidator {
    private static final String DEFAULT_URL = "https://api.typesafe.ai";
    // Keep these questions in sync with internal/rules/catalog.json.
    private static final Map<String, String> RULE_QUESTIONS = Map.of(
            "person_name", "¿Este texto parece representar el nombre de una persona?",
            "product_description", "¿El texto identifica un producto y al menos una característica concreta, más allá de una opinión genérica?",
            "address", "¿El texto parece una dirección física suficientemente específica para ubicar un lugar, y no solo el nombre de una ciudad o país?");

    private final String key;
    private final String baseUrl;
    private final Duration timeout;
    private final HttpClient client;
    private final ObjectMapper mapper = new ObjectMapper();

    public DirectValidator(String jevApiKey) {
        this(jevApiKey, DEFAULT_URL, Duration.ofSeconds(30));
    }

    public DirectValidator(String jevApiKey, String baseUrl, Duration timeout) {
        if (jevApiKey == null || jevApiKey.isBlank()) throw new IllegalArgumentException("jevApiKey is required");
        if (timeout == null || timeout.isZero() || timeout.isNegative()) throw new IllegalArgumentException("timeout must be positive");
        URI endpoint = URI.create(baseUrl);
        String host = endpoint.getHost();
        boolean official = "https".equalsIgnoreCase(endpoint.getScheme()) && "api.typesafe.ai".equals(host);
        boolean loopback = "http".equalsIgnoreCase(endpoint.getScheme()) &&
                ("localhost".equals(host) || "127.0.0.1".equals(host) || "[::1]".equals(host));
        if (host == null || (!official && !loopback) ||
                endpoint.getUserInfo() != null || endpoint.getQuery() != null || endpoint.getFragment() != null ||
                !(endpoint.getPath().isEmpty() || "/".equals(endpoint.getPath()))) {
            throw new IllegalArgumentException("Jev endpoint must be api.typesafe.ai (except loopback tests)");
        }
        this.key = jevApiKey.trim();
        this.baseUrl = baseUrl.replaceAll("/+$", "");
        this.timeout = timeout;
        this.client = HttpClient.newBuilder().connectTimeout(timeout).followRedirects(HttpClient.Redirect.NEVER).build();
    }

    public ValidationResult validate(String rule, String value) throws IOException, InterruptedException {
        validText(value, "value");
        String question = RULE_QUESTIONS.get(rule);
        if (question == null) throw new SemanticValidatorException("requested rule is not registered", "unknown_rule", 0);
        return check(value, question);
    }

    public ValidationResult name(String value) throws IOException, InterruptedException {
        return validate("person_name", value);
    }

    public ValidationResult check(String value, String question) throws IOException, InterruptedException {
        String state = validText(value, "value");
        String instructions = validText(question, "question");
        String body = mapper.writeValueAsString(Map.of("model", "jev-latest", "state", state,
                "questions", Map.of("result", Map.of("type", "choice", "instructions", instructions,
                        "criteria", Map.of("true", "El texto cumple la condición.",
                                "false", "El texto no cumple la condición.")))));
        HttpRequest request = HttpRequest.newBuilder(URI.create(baseUrl + "/v1/systemone"))
                .timeout(timeout)
                .header("Accept", "application/json")
                .header("Authorization", "Bearer " + key)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(body)).build();
        HttpResponse<String> response;
        try {
            response = client.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (java.net.http.HttpTimeoutException exception) {
            throw new SemanticValidatorException("Jev request timed out", "provider_timeout", 0);
        } catch (IOException exception) {
            throw new SemanticValidatorException("Jev request failed", "provider_error", 0);
        }
        if (response.statusCode() < 200 || response.statusCode() >= 300) {
            throw new SemanticValidatorException("Jev returned HTTP " + response.statusCode(), "provider_error", response.statusCode());
        }
        if (response.body().length() > 2 * 1024 * 1024) throw invalidResponse();
        JsonNode answer;
        try {
            answer = mapper.readTree(response.body()).path("answers").path("result");
        } catch (IOException | NullPointerException exception) {
            throw invalidResponse();
        }
        JsonNode choiceNode = answer.path("choice");
        String choice = choiceNode.isTextual() ? choiceNode.asText() : "";
        if (!"true".equals(choice) && !"false".equals(choice)) throw invalidResponse();
        JsonNode confidenceNode = answer.path("confidence");
        if (confidenceNode.isMissingNode()) confidenceNode = answer.path("probabilities").path(choice);
        if (!confidenceNode.isNumber()) throw invalidResponse();
        double confidence = confidenceNode.asDouble();
        if (!Double.isFinite(confidence) || confidence < 0 || confidence > 1) throw invalidResponse();
        if (confidence < 0.8) return new ValidationResult(null, confidence, "uncertain");
        boolean valid = "true".equals(choice);
        return new ValidationResult(valid, confidence, valid ? "valid" : "invalid");
    }

    public List<BatchResult> validateBatch(List<BatchItem> items) throws InterruptedException {
        if (items == null || items.isEmpty() || items.size() > 100) throw new IllegalArgumentException("items must contain between 1 and 100 elements");
        Set<String> ids = new HashSet<>();
        for (BatchItem item : items) {
            if (item == null || item.getId() == null || item.getId().isBlank() || !ids.add(item.getId())) {
                throw new IllegalArgumentException("batch items require unique ids");
            }
        }
        ExecutorService executor = Executors.newFixedThreadPool(Math.min(4, items.size()));
        try {
            List<Callable<BatchResult>> jobs = new ArrayList<>();
            for (BatchItem item : items) jobs.add(() -> {
                try {
                    ValidationResult result = validate(item.getRule(), item.getValue());
                    return new BatchResult(item.getId(), result.getValid(), result.getConfidence(), result.getStatus(), null);
                } catch (IOException exception) {
                    String code = exception instanceof SemanticValidatorException ? ((SemanticValidatorException) exception).getCode() : "provider_error";
                    String message = "unknown_rule".equals(code) ? "requested rule is not registered" :
                            "invalid_request".equals(code) ? "invalid value" : "semantic provider request failed";
                    return new BatchResult(item.getId(), null, 0, "error", new BatchItemError(code, message));
                }
            });
            List<BatchResult> results = new ArrayList<>();
            for (Future<BatchResult> future : executor.invokeAll(jobs)) {
                try { results.add(future.get()); }
                catch (ExecutionException exception) { throw new IllegalStateException("batch worker failed", exception); }
            }
            return results;
        } finally {
            executor.shutdownNow();
        }
    }

    private static String validText(String text, String field) throws SemanticValidatorException {
        if (text == null || text.isBlank() || text.trim().codePointCount(0, text.trim().length()) > 10000) {
            throw new SemanticValidatorException(field + " must contain between 1 and 10000 characters", "invalid_request", 0);
        }
        return text.trim();
    }

    private static SemanticValidatorException invalidResponse() {
        return new SemanticValidatorException("invalid Jev response", "invalid_response", 0);
    }
}
