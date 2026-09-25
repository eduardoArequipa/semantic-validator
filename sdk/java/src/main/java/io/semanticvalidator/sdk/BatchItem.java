package io.semanticvalidator.sdk;

public class BatchItem {
    private final String id;
    private final String rule;
    private final String value;

    public BatchItem(String id, String rule, String value) {
        this.id = id;
        this.rule = rule;
        this.value = value;
    }

    public String getId() {
        return id;
    }

    public String getRule() {
        return rule;
    }

    public String getValue() {
        return value;
    }
}
