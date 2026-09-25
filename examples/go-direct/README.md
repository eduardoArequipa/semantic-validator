# Go: direct SDK example

From the repository root, run the offline tests:

```bash
go test ./examples/go-direct
```

They start a fake Jev server on the local loopback interface, use the dummy key
`test-only`, and check a positive decision, an uncertain result, and a provider
error. They do not contact Jev or consume credits.

For a live run, make your own Jev key available as `TYPESAFE_API_KEY` to your
backend process and run:

```bash
go run ./examples/go-direct
```

This sends two requests to Jev and may consume credits. Do not put the key in
source code or a browser. The direct SDK comes from this repository's Go
module; no Semantic Validator server is required.
