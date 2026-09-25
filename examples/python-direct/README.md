# Python: direct SDK example

From the repository root, create a virtual environment and install the SDK
from this checkout:

```bash
python3 -m venv examples/python-direct/.venv
examples/python-direct/.venv/bin/python -m pip install -e sdk/python
examples/python-direct/.venv/bin/python -m unittest discover -s examples/python-direct -p 'test_*.py'
```

The tests use `test-only` and a fake HTTP response. They do not use a real Jev
key, send network requests, or consume Jev credits. They exercise a positive
decision, an uncertain result, and a provider error.

To run the same program against Jev, make your own key available as
`TYPESAFE_API_KEY` to the process, then run:

```bash
examples/python-direct/.venv/bin/python examples/python-direct/main.py
```

This makes two real Jev requests, which may consume credits. Keep the key in a
trusted backend and out of source control. A `.env` file is not loaded
automatically. `uncertain` is not `false`; a provider error is neither one.
