# Semantic Validator v0.3.0

First tagged beta release of the free, open-source Semantic Validator SDKs.
Each SDK can call Jev directly using **your own Jev key** from a trusted backend;
the optional Go REST server can be self-hosted. Our public domain serves docs
only and does not accept validation requests.

Includes TypeScript, Python, Go, and Java SDK source packages, the three
built-in field rules (`person_name`, `product_description`, `address`), and a
TypeScript form example. See the [SDK guide](https://validator.trialsur.cloud/docs/byok-en.html)
and [installation instructions](https://github.com/eduardoArequipa/semantic-validator/blob/v0.3.0/docs/RELEASE.md). Verify downloads against
`SHA256SUMS.txt` before using them.

The [90-case evaluation](https://github.com/eduardoArequipa/semantic-validator/blob/v0.3.0/evals/RESULTS-2026-09-25.md) is a small fictional
benchmark, **not** a real-world accuracy claim. Semantic results can vary;
handle `uncertain` and provider errors separately. Never expose your Jev key
in browser code or commit it to a repository.

Jev is a separate service, not included in this release. Jev usage may incur
charges under your own TypeSafe agreement.
