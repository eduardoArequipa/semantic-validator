package io.semanticvalidator.sdk;

public class BatchItemError {
    private final String code;
    private final String message;

    public BatchItemError(String code, String message) {
        this.code = code;
        this.message = message;
    }

    public String getCode() {
        return code;
    }

    public String getMessage() {
        return message;
    }
}
