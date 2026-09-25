# Evaluate Semantic Validator

[Español](README.es.md) · [Project README](../README.md)

For the separate 100-case Jev–Laya experiment, see
[Provider comparison](PROVIDER_COMPARISON.md). It does not change production.

This evaluator compares **direct Jev or your own server** responses with examples
labeled before the run. The original ten fictional examples cover Spanish
and English. [`field-rules.json`](field-rules.json) adds 30 fictional cases
per built-in rule (12 positive, 12 negative, 6 ambiguous). These are starter
regression examples, not a claim of measured accuracy. Each live request
may use Jev credits and, in individual-key mode, one local quota unit.
A dry run sends none:

```bash
python3 tools/evaluate.py --dry-run
python3 tools/evaluate.py --cases evals/field-rules.json --dry-run
```

The [nine-case pilot](PILOT-2026-09-25.md) is documented separately; it is
not a production-quality accuracy estimate.
The [full 90-case run](RESULTS-2026-09-25.md) publishes aggregate results,
methodology, and limitations without publishing a key or private text.

For an application-specific evaluation, copy `evals/cases.example.json` to
`evals/private/cases.json`. This private directory is excluded from Git. Each
case needs a non-sensitive `id`, `value`, `expected_valid` (`true`, `false`, or
`null` for intentionally ambiguous cases),
and exactly one of `rule` or `question`. Label the expected answer first and
use the same question for the examples you compare.

The recommended live path calls Jev directly with your own account. Install
the Python SDK in a virtual environment and enter your Jev key without putting
it in shell history. A full 90-case run can consume **90 Jev calls**; use
`--sample-per-label 1` for a balanced first run of nine calls (one positive,
one negative, and one ambiguous case per rule):

```bash
python -m pip install -e sdk/python
read -rsp 'Jev key: ' TYPESAFE_API_KEY
printf '\n'
export TYPESAFE_API_KEY
python3 tools/evaluate.py --direct --cases evals/field-rules.json --sample-per-label 1
unset TYPESAFE_API_KEY
```

The hosted API is paused. To evaluate through your own Go server instead,
run it with your Jev key as described in [Getting Started](../docs/GETTING_STARTED.md),
then use the local development key from `.env.example` or your own instance key:

```bash
SEMANTIC_VALIDATOR_API_KEY=local-dev python3 tools/evaluate.py \
  --cases evals/field-rules.json --base-url http://localhost:8080
```

The report shows case IDs, expected answers, returned status and confidence,
false positives, false negatives, uncertain results, ambiguous outcomes and
API errors, overall and per rule. It never
prints the key or example texts and does not save responses. An HTTP failure
is not counted as a semantic decision. Repeat the run after changing a rule
or provider, using the same labeled set.

Historically, we ran the ten original fictional examples against the hosted
API on 2026-09-23 (before it was paused):
**9 matched their labels, 1 was uncertain, and 0 returned API errors**. The
uncertain case was an English vague support request. This sample is not a
claim of equal quality across languages. Add many more examples from the
actual product domain before making such a claim.

Do not publish private texts or send them to Jev unless they are
authorized for processing by Jev. To test the evaluator locally without
making real requests:

```bash
python3 -m unittest discover -s tools -p 'test_*.py'
```
