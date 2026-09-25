package io.semanticvalidator.sdk;

import java.io.IOException;

public class SemanticValidatorException extends IOException {
    private final String code;
    private final int statusCode;

    public SemanticValidatorException(String message, String code, int statusCode) {
        super(message);
        this.code = code;
        this.statusCode = statusCode;
    }

    public String getCode() {
        return code;
    }

    public int getStatusCode() {
        return statusCode;
    }
}
