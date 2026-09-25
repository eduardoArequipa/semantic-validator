# Install a tagged release

Start with the [v0.3.0 release page](https://github.com/eduardoArequipa/semantic-validator/releases/tag/v0.3.0).
Download only the SDK you need plus `SHA256SUMS.txt`. Verify it in the same
directory with `sha256sum -c SHA256SUMS.txt` after downloading all four SDK
archives, or compare the single archive's digest with its line in that file.
Never put your Jev key in a frontend, app bundle, or repository.

## TypeScript / Node.js

Requires Node.js 18+ in a trusted backend. Download
`semantic-validator-sdk-0.3.0.tgz` and run:

```bash
npm install ./semantic-validator-sdk-0.3.0.tgz
```

Import `DirectValidator` from `@semantic-validator/sdk`. See the
[TypeScript guide](../sdk/typescript/README.md) for a complete example.
The package is not yet on npm.

## Python

Requires Python 3.10+ and a virtual environment. Download
`semantic-validator-python-0.3.0.tar.gz`, then:

```bash
tar -xzf semantic-validator-python-0.3.0.tar.gz
python3 -m venv .venv
. .venv/bin/activate
python -m pip install ./semantic-validator/sdk/python
```

Import `DirectValidator` from `semantic_validator`. See the
[Python guide](../sdk/python/README.md). The package is not yet on PyPI.

## Go

The Go SDK is part of the tagged root module:

```bash
go get github.com/eduardoArequipa/semantic-validator/sdk/go@v0.3.0
```

Import `github.com/eduardoArequipa/semantic-validator/sdk/go` and call
`validator.NewDirectClient` (using your own key). See the
[Go guide](../sdk/go/README.md). The release also includes an optional Go SDK
source archive.

## Java

Requires Java 11+ and Maven. Download
`semantic-validator-java-0.3.0.tar.gz`, then:

```bash
tar -xzf semantic-validator-java-0.3.0.tar.gz
mvn -f semantic-validator/sdk/java/pom.xml install
```

Use the coordinates `io.semanticvalidator:semantic-validator-sdk:0.3.0` in
your own Maven project. This installs the SDK in your **local** Maven cache;
it is not yet on Maven Central. See the [Java guide](../sdk/java/README.md).

## Before using results in production

Use `TYPESAFE_API_KEY` only in a trusted backend. The direct SDKs call Jev
with your account; the hosted Semantic Validator API is paused. A semantic
answer can be `uncertain` and must not be treated as a negative result.
Provider failures are errors, not negative results. The [evaluation](../evals/RESULTS-2026-09-25.md)
uses a small fictional dataset and does not establish real-world accuracy.
