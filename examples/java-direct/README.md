# Java: direct SDK example

Java 11+ and Maven are required. From the repository root, first install the
SDK into your local Maven cache, then run the example tests:

```bash
mvn -B -f sdk/java/pom.xml install
mvn -B -f examples/java-direct/pom.xml test
```

The tests use the dummy key `test-only` and a local fake Jev server. They do
not contact Jev or consume credits. They verify a positive decision, an
uncertain result, and a provider error.

For a live run, make your own Jev key available as `TYPESAFE_API_KEY` to your
backend process, then:

```bash
mvn -B -f examples/java-direct/pom.xml compile exec:java
```

This sends two real Jev requests and may consume credits. The key belongs in
your environment or secret manager, never in the Java source or browser. The
example uses your account directly; our hosted API is not involved.
