package io.semanticvalidator.sdk;

public class ValidationResult {
    private final Boolean valid;
    private final double confidence;
    private final String status;

    public ValidationResult(Boolean valid, double confidence, String status) {
        this.valid = valid;
        this.confidence = confidence;
        this.status = status;
    }

    public Boolean getValid() {
        return valid;
    }

    public double getConfidence() {
        return confidence;
    }

    public String getStatus() {
        return status;
    }
}
