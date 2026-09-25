# Contributing

[Español](CONTRIBUTING.es.md) · [README](README.md)

Thanks for helping improve Semantic Validator. The source is licensed under
[Apache License 2.0](LICENSE). Contributions are welcome in the
[GitHub repository](https://github.com/eduardoArequipa/semantic-validator).

## Development setup

- Go 1.22+: run `go test ./...` and `go vet ./...`.
- Python 3.10+: run `python3 -m unittest discover -s tools -p 'test_*.py'`.
- Node.js 22+: run `npm ci` and `npm test` in both `sdk/typescript` and
  `examples/typescript-demo`.
- Go tests also verify the embedded SDK documentation and downloads.
- Build the Java SDK with `mvn -f sdk/java/pom.xml package`.

The automated tests use simulated responses and do not call Jev. Do not commit
`.env`, API keys, customer text, or private evaluation cases. See
[the evaluation guide](evals/README.md) to compare semantic behavior with
examples labeled before the run.

Describe the use case, expected behavior, and a reproducing test with each
change. For security issues, do not post secrets or customer data in public
issues; contact the deployment administrator privately.
