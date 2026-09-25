package io.semanticvalidator.sdk;

public class BatchResult extends ValidationResult {
    private final String id;
    private final BatchItemError error;

    public BatchResult(String id, Boolean valid, double confidence, String status, BatchItemError error) {
        super(valid, confidence, status);
        this.id = id;
        this.error = error;
    }

    public String getId() {
        return id;
    }

    public BatchItemError getError() {
        return error;
    }
}
