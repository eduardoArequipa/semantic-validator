# Try every SDK without a Jev key

Clone the repository and run the examples from its root:

```bash
git clone https://github.com/eduardoArequipa/semantic-validator.git
cd semantic-validator
```

Each directory has a program for your **own Jev key** and an offline test with
simulated responses. Start with the test: it never contacts Jev or consumes
credits. A live run requires `TYPESAFE_API_KEY` in a trusted backend process;
never put the key in source code or a browser. A `.env` file is not loaded
automatically, except by the TypeScript demo's explicit run scripts.

| Language | No-key test | Live run |
| --- | --- | --- |
| [TypeScript](typescript-demo/README.md) | `cd examples/typescript-demo && npm ci && npm test` | In that directory: `npm run prueba` (reads `.env`) |
| [Python](python-direct/README.md) | `PYTHONPATH=sdk/python/src python3 -m unittest discover -s examples/python-direct -p 'test_*.py'` | `python3 examples/python-direct/main.py` (install the SDK first) |
| [Go](go-direct/README.md) | `go test ./examples/go-direct` | `go run ./examples/go-direct` |
| [Java](java-direct/README.md) | `mvn -f sdk/java/pom.xml install && mvn -f examples/java-direct/pom.xml test` | `mvn -f examples/java-direct/pom.xml compile exec:java` |

For Python, follow its README to create a virtual environment. Java installs
the SDK into your local Maven cache before testing the example. Go and Java
tests bind **only to a local fake server** on the loopback interface; no real credential
is needed. TypeScript already has mock tests for complaints and forms.

Live runs make Jev requests and may consume credits. `uncertain` needs review;
a provider error is not `false`. See the [usage guide](https://validator.trialsur.cloud/docs/byok-en.html#usage)
for result handling.
