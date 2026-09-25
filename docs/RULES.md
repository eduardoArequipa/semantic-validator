# Built-in field rules (beta)

[Español](RULES.es.md) · [Evaluation guide](../evals/README.md)

The versioned source of truth is [`internal/rules/catalog.json`](../internal/rules/catalog.json).
`/v1/validate` on your own server and `validate(rule, value)` in the direct SDKs
use the same questions. Changing a question requires a new rule version so cached
results and evaluations are not silently reused. The server includes the version
in its cache key. Direct SDK calls do not use our cache.

| Rule (v1) | Intended yes | Intended no | Review examples |
| --- | --- | --- | --- |
| `person_name` | Plausible personal name | Random text, company, product or role | One-word names also used as ordinary words |
| `product_description` | Identifies a product and at least one concrete feature | Generic praise or a sales slogan alone | Product name with a vague or missing feature |
| `address` | Specific physical location | Country or city alone, unrelated text | Landmark-only or incomplete street directions |

These rules ask whether text *appears* to meet the description. They do not
verify a person's identity, a product's factual specifications, or whether an
address exists or is deliverable. Empty/oversized input is a request error;
`invalid` is a semantic decision, `uncertain` means review, and a provider
failure is an error. A Jev confidence score is **not** a calibrated probability
of correctness. Applications should decide how to handle `uncertain` and
errors before using these rules to block users.

The public [`field-rules.json`](../evals/field-rules.json) contains 30 fictional
cases per rule: 12 positive, 12 negative, and 6 intentionally ambiguous. The
ambiguous label `null` is tracked for review, not scored as right or wrong.
These hand-written examples are a starting regression set, not evidence of
production accuracy. Run a live evaluation with your own Jev key and review
false positives, false negatives and uncertain cases before relying on a rule.
