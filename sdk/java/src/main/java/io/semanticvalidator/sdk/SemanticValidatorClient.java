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
import java.util.Objects;
import java.util.Set;

public class SemanticValidatorClient {
    private static final String DEFAULT_BASE_URL = "http://localhost:8080";
    private static final Duration DEFAULT_TIMEOUT = Duration.ofSeconds(30);

    private final String apiKey;
    private final String baseUrl;
    private final Duration timeout;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public SemanticValidatorClient(String apiKey) {
        this(apiKey, DEFAULT_BASE_URL, DEFAULT_TIMEOUT);
    }

    public SemanticValidatorClient(String apiKey, String baseUrl, Duration timeout) {
        if (apiKey == null || apiKey.isBlank()) {
            throw new IllegalArgumentException("apiKey is required");
        }
        if (baseUrl == null || baseUrl.isBlank()) {
            throw new IllegalArgumentException("baseUrl is required");
        }
        if (timeout == null || timeout.isZero() || timeout.isNegative()) {
            throw new IllegalArgumentException("timeout must be positive");
        }

        String normalizedBaseUrl = baseUrl.replaceAll("/+$", "");
        URI baseUri = URI.create(normalizedBaseUrl);
        if (!("http".equalsIgnoreCase(baseUri.getScheme()) || "https".equalsIgnoreCase(baseUri.getScheme()))
                || baseUri.getHost() == null) {
            throw new IllegalArgumentException("baseUrl must be an absolute HTTP or HTTPS URL");
        }

        this.apiKey = apiKey;
        this.baseUrl = normalizedBaseUrl;
        this.timeout = timeout;
        this.httpClient = HttpClient.newBuilder().connectTimeout(timeout).build();
        this.objectMapper = new ObjectMapper();
    }

    public ValidationResult validate(String rule, String value) throws IOException, InterruptedException {
        Objects.requireNonNull(rule, "rule is required");
        Objects.requireNonNull(value, "value is required");
        JsonNode payload = post("/v1/validate", Map.of("rule", rule, "value", value));
        return parseValidationResult(payload, false);
    }

    public ValidationResult name(String value) throws IOException, InterruptedException {
        return validate("person_name", value);
    }

    public ValidationResult check(String value, String question) throws IOException, InterruptedException {
        Objects.requireNonNull(value, "value is required");
        Objects.requireNonNull(question, "question is required");
        JsonNode payload = post("/v1/check", Map.of("value", value, "question", question));
        return parseValidationResult(payload, false);
    }

    public List<BatchResult> validateBatch(List<BatchItem> items) throws IOException, InterruptedException {
        Objects.requireNonNull(items, "items are required");
        if (items.isEmpty() || items.size() > 100) {
            throw new IllegalArgumentException("items must contain between 1 and 100 elements");
        }
        Set<String> ids = new HashSet<>();
        for (BatchItem item : items) {
            Objects.requireNonNull(item, "batch item is required");
            if (item.getId() == null || item.getId().isEmpty()) {
                throw new IllegalArgumentException("each batch item requires an id");
            }
            if (!ids.add(item.getId())) {
                throw new IllegalArgumentException("batch item ids must be unique: " + item.getId());
            }
            Objects.requireNonNull(item.getRule(), "batch item rule is required");
            Objects.requireNonNull(item.getValue(), "batch item value is required");
        }

        JsonNode payload = post("/v1/validate/batch", Map.of("items", items));
        JsonNode resultItems = payload.path("items");
        if (!resultItems.isArray() || resultItems.size() != items.size()) {
            throw invalidResponse("server returned an invalid batch response");
        }

        List<BatchResult> results = new ArrayList<>(resultItems.size());
        for (int i = 0; i < resultItems.size(); i++) {
            JsonNode item = resultItems.get(i);
            String id = textField(item, "id");
            if (!items.get(i).getId().equals(id)) {
                throw invalidResponse("server returned batch results in an unexpected order");
            }
            ValidationResult result = parseValidationResult(item, true);
            BatchItemError error = null;
            JsonNode errorNode = item.get("error");
            if (errorNode != null && !errorNode.isNull()) {
                String code = textField(errorNode, "code");
                String message = textField(errorNode, "message");
                error = new BatchItemError(code, message);
            }
            if ("error".equals(result.getStatus()) && error == null) {
                throw invalidResponse("batch error item is missing error details");
            }
            results.add(new BatchResult(id, result.getValid(), result.getConfidence(), result.getStatus(), error));
        }
        return results;
    }

    private JsonNode post(String path, Object payload) throws IOException, InterruptedException {
        String requestBody = objectMapper.writeValueAsString(payload);
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .timeout(timeout)
                .header("Accept", "application/json")
                .header("Authorization", "Bearer " + apiKey)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(requestBody))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        JsonNode responseBody;
        try {
            responseBody = objectMapper.readTree(response.body());
        } catch (IOException exception) {
            throw new SemanticValidatorException("server returned invalid JSON", "invalid_response", response.statusCode());
        }
        if (responseBody == null) {
            throw new SemanticValidatorException("server returned an empty response", "invalid_response", response.statusCode());
        }
        if (response.statusCode() < 200 || response.statusCode() >= 300) {
            JsonNode error = responseBody.path("error");
            String code = optionalText(error, "code", "http_error");
            String message = optionalText(error, "message", "request failed");
            throw new SemanticValidatorException(message, code, response.statusCode());
        }
        return responseBody;
    }

    private static ValidationResult parseValidationResult(JsonNode payload, boolean allowBatchError)
            throws SemanticValidatorException {
        if (payload == null || !payload.isObject()) {
            throw invalidResponse("server returned an invalid validation response");
        }
        JsonNode validNode = payload.get("valid");
        JsonNode confidenceNode = payload.get("confidence");
        String status = textField(payload, "status");
        boolean validStatus = "valid".equals(status) || "invalid".equals(status) || "uncertain".equals(status);
        boolean errorStatus = allowBatchError && "error".equals(status);
        if (validNode == null || !(validNode.isBoolean() || validNode.isNull())
                || confidenceNode == null || !confidenceNode.isNumber()
                || (!validStatus && !errorStatus)) {
            throw invalidResponse("server returned an invalid validation response");
        }
        double confidence = confidenceNode.asDouble();
        if (!Double.isFinite(confidence) || confidence < 0 || confidence > 1) {
            throw invalidResponse("server returned an invalid confidence value");
        }
        Boolean valid = validNode.isNull() ? null : validNode.asBoolean();
        return new ValidationResult(valid, confidence, status);
    }

    private static String textField(JsonNode object, String field) throws SemanticValidatorException {
        JsonNode value = object == null ? null : object.get(field);
        if (value == null || !value.isTextual()) {
            throw invalidResponse("server returned an invalid " + field + " field");
        }
        return value.asText();
    }

    private static String optionalText(JsonNode object, String field, String fallback) {
        JsonNode value = object == null ? null : object.get(field);
        return value != null && value.isTextual() ? value.asText() : fallback;
    }

    private static SemanticValidatorException invalidResponse(String message) {
        return new SemanticValidatorException(message, "invalid_response", 0);
    }
}
