# Evaluate Semantic Validator

[Español](README.es.md) · [Project README](../README.md)

For the separate 100-case Jev–Laya experiment, see
[Provider comparison](PROVIDER_COMPARISON.md). It does not change production.

This evaluator compares API responses with examples labeled by a person
before the run. The included ten fictional examples cover Spanish and English.
They are illustrative and too few to estimate overall accuracy. Each live
request may use one quota unit, including cache hits. A dry run sends none:

```bash
python3 tools/evaluate.py --dry-run
```

For an application-specific evaluation, copy `evals/cases.example.json` to
`evals/private/cases.json`. This private directory is excluded from Git. Each
case needs a non-sensitive `id`, `value`, `expected_valid` (`true` or `false`),
and exactly one of `rule` or `question`. Label the expected answer first and
use the same question for the examples you compare.

Enter an individual Semantic Validator key in Bash without adding it to your
shell history:

```bash
read -rsp 'API key: ' SEMANTIC_VALIDATOR_API_KEY
printf '\n'
export SEMANTIC_VALIDATOR_API_KEY
python3 tools/evaluate.py --cases evals/private/cases.json \
  --base-url https://validator.trialsur.cloud
unset SEMANTIC_VALIDATOR_API_KEY
```

The report shows case IDs, expected answers, returned status and confidence,
false positives, false negatives, uncertain results, and API errors. It never
prints the key or example texts and does not save responses. An HTTP failure
is not counted as a semantic decision. Repeat the run after changing a rule
or provider, using the same labeled set.

We ran the ten fictional examples against the hosted API on 2026-09-23:
**9 matched their labels, 1 was uncertain, and 0 returned API errors**. The
uncertain case was an English vague support request. This sample is not a
claim of equal quality across languages. Add many more examples from the
actual product domain before making such a claim.

Do not publish private texts or send them to this API unless they are
authorized for processing by Jev. To test the evaluator locally without
making real requests:

```bash
python3 -m unittest discover -s tools -p 'test_*.py'
```
