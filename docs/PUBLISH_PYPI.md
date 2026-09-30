# Publish the Python SDK to PyPI

Package name: `semantic-validator`. Import name: `semantic_validator`.
The first PyPI release is version `0.3.0`.

## Prepare the distributions

From the repository root:

```bash
python3 -m venv .venv-publish
.venv-publish/bin/python -m pip install build twine 'setuptools>=77.0.3' wheel
.venv-publish/bin/python -m build sdk/python --outdir dist/pypi --no-isolation
.venv-publish/bin/python -m twine check \
  dist/pypi/semantic_validator-0.3.0-py3-none-any.whl \
  dist/pypi/semantic_validator-0.3.0.tar.gz
```

The wheel and source distribution must contain the SDK, Apache-2.0 license,
and valid project metadata. Install the wheel in a separate virtual environment
and run the SDK tests before uploading. These tests use simulated provider
responses and require no Jev credentials.

## First upload

Create a [PyPI account](https://pypi.org/account/register/), verify the email,
and enable two-factor authentication. PyPI supports an authenticator application
(TOTP) as well as security devices. Save the account recovery codes.

In account settings, create an API token. Since the project does not exist
yet, use the `Entire account` scope for the first upload. Keep the token local;
do not paste it into source code, Git, shell command arguments, or issue reports.

```bash
.venv-publish/bin/python -m twine upload \
  --repository-url https://upload.pypi.org/legacy/ \
  --username __token__ \
  dist/pypi/semantic_validator-0.3.0-py3-none-any.whl \
  dist/pypi/semantic_validator-0.3.0.tar.gz
```

Paste the PyPI API token when Twine prompts for it. The input is hidden.
The Jev key is not needed to publish.

After the upload, verify [the project page](https://pypi.org/project/semantic-validator/0.3.0/)
and install the exact version in a fresh virtual environment:

```bash
python3 -m venv /tmp/semantic-validator-pypi-verification
/tmp/semantic-validator-pypi-verification/bin/python -m pip install \
  --index-url https://pypi.org/simple/ semantic-validator==0.3.0
/tmp/semantic-validator-pypi-verification/bin/python -c \
  'from semantic_validator import DirectValidator, Validator; print("SDK imports successfully")'
```

Create a project-scoped token for future manual releases and revoke the
account-wide token used for this first upload. Future versions must change
`version` in `sdk/python/pyproject.toml`; PyPI does not let you overwrite a
published file. A future GitHub Actions publisher can use Trusted Publishing
instead of a stored token.

References: [Python packaging tutorial](https://packaging.python.org/en/latest/tutorials/packaging-projects/),
[PyPI account and token help](https://pypi.org/help/),
[Trusted Publishing](https://docs.pypi.org/trusted-publishers/).
