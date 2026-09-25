# Jev–Laya provider experiment

This experiment leaves the public API and production Jev provider unchanged.
It compares the **current Semantic Validator product path** (hosted API backed
by Jev) with a **local Laya server** on 100 explicitly labeled, fictional cases.
The set has five scenarios (`person_name`, complaints, purchase intent, product
descriptions, failure reports), 50 Spanish and 50 English cases, and 50 positive
and 50 negative labels. The dataset is synthetic, manually drafted, and not an
estimate of real-world accuracy. Labels should be independently reviewed before
using results to make a product claim.

## Prepare Laya locally

Laya is not a dependency of the Go service. Follow the [official Laya
installation guide](https://github.com/NandhaKishorM/laya#installation-details)
for your CPU/GPU and Python version. One possible CPU setup in a separate
virtual environment is:

```bash
python3 -m venv .venv-laya
.venv-laya/bin/python -m pip install --upgrade pip
.venv-laya/bin/python -m pip install 'torch==2.14.0+cpu' \
  --index-url https://download.pytorch.org/whl/cpu \
  --extra-index-url https://pypi.org/simple
.venv-laya/bin/python -m pip install 'laya[serve]'
LAYA_HOST=127.0.0.1 LAYA_PORT=8000 LAYA_DEVICE=cpu \
  LAYA_PRELOAD=1 LAYA_MODELS=english,multilingual .venv-laya/bin/laya-serve
```

Wait for startup to complete: preloading downloads both model checkpoints on
the first run. Do **not** expose its default unauthenticated HTTP server on a
public interface. Consult the upstream guide if local hardware or package
versions require different setup.

## Run the comparison

First inspect and validate the cases without any network calls:

```bash
python3 tools/compare_providers.py --dry-run
python3 -m unittest discover -s tools -p 'test_*.py'
```

The live run uses a Semantic Validator API key and at most one hosted API call
plus one local Laya call per case. Allow enough quota (100 hosted validations
for the full set), and avoid running other hosted requests at the same time if
you want cleaner latency observations. Enter the key without adding it to
shell history:

```bash
read -rsp 'Semantic Validator API key: ' SEMANTIC_VALIDATOR_API_KEY
printf '\n'
export SEMANTIC_VALIDATOR_API_KEY
python3 tools/compare_providers.py --failures-only
unset SEMANTIC_VALIDATOR_API_KEY
```

For a smoke test, add `--limit 4`. `--jev-base-url` and `--laya-base-url` may be
changed to other trusted endpoints; plain HTTP is accepted only for loopback.
If the local server has `LAYA_API_KEY` configured, export the same value for the
comparison process. Never point a server at an untrusted URL while supplying
an API key. The script does not print or save texts, keys, or raw responses;
it prints case IDs, aggregate outcomes by language/scenario, and median observed
request latency. It stops on a hosted 429 quota/rate-limit response.

## Interpret the results

The hosted path uses the current `choice` request and 0.80 uncertainty cutoff.
Laya uses a `noul` question with neutral model-facing labels `A`/`B`, then
applies the same 0.80 cutoff to `max(P(true), 1-P(true))`. This attempts to
reduce Laya's documented `true`/`false` label sensitivity, but is **not a controlled
model-to-model benchmark**: the request primitives and infrastructure differ.
For a local-only sensitivity check, use `--laya-only --laya-primitive choice`.
This sends `choice` options `A`/`B` and gates on Laya's `answer_confidence`,
not its entropy-based `confidence` field. It consumes no Semantic Validator
quota and needs no hosted API key.
Laya can still be sensitive to wording and negation. Inspect failures by case
ID and confirm the labels before drawing conclusions. A server-side cache hit
or first Laya model load makes the reported latency incomparable as pure model
speed; warm both paths and repeat if latency matters.

The `person_name` Laya question is a frozen copy of rule v1 from
`internal/rules/registry.go`; update it if the production rule version changes.
Do not set a universal pass/fail threshold from this small synthetic dataset.
Only consider an experimental provider after testing authorized, realistic
examples and checking calibration separately for Spanish and English.

## First live run: 2026-09-24

The complete 100-case fictional set was run sequentially against the hosted
Semantic Validator API and local Laya 0.3.20 on CPU. Both Laya checkpoints
(English and multilingual) loaded. Results at the 0.80 cutoff:

| Path | Correct | False positives | False negatives | Uncertain | HTTP errors |
| --- | ---: | ---: | ---: | ---: | ---: |
| Hosted API / Jev | 93 | 0 | 0 | 7 | 0 |
| Local Laya / `noul` | 17 | 5 | 0 | 78 | 0 |
| Local Laya / `choice` A/B (second, local-only run) | 17 | 10 | 5 | 68 | 0 |

Jev resolved 46/50 Spanish and 47/50 English examples; Laya `noul` resolved
9/50 Spanish and 8/50 English examples correctly. Observed median request
latencies were 1,296 ms for the hosted Jev path, 696 ms for local Laya `noul`,
and 881 ms for local Laya `choice`. These are **not model-only speed numbers**:
the hosted path includes network, the paths may have cache effects, and the
Laya variants were run at different times. The English Laya checkpoint emitted
a temperature warning for a `choice` bucket during warmup; that warning does
not by itself establish the calibration of these particular yes/no questions.

These figures describe this exact prompt format, checkpoints, hand-labeled
synthetic set, and threshold. In particular, the base Laya checkpoints have
not been fine-tuned for our tasks. The `noul` and `choice` results show that
merely changing the two-option format did not fix quality here. Do not claim
Jev's 93% as general accuracy or offer Laya as a drop-in production provider
from these data. The next useful experiment is an independently reviewed,
authorized real-world set and, if Laya remains of interest, domain-specific
fine-tuning plus calibration. No production configuration changed.
