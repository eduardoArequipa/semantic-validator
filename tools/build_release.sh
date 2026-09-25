#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 || ! $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Usage: tools/build_release.sh vMAJOR.MINOR.PATCH" >&2
  exit 2
fi

version=${1#v}
repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

python3 - "$version" <<'PY'
import json
import pathlib
import sys
import tomllib
import xml.etree.ElementTree as ET

version = sys.argv[1]
root = pathlib.Path.cwd()
versions = {
    "TypeScript": json.loads((root / "sdk/typescript/package.json").read_text())["version"],
    "Python": tomllib.loads((root / "sdk/python/pyproject.toml").read_text())["project"]["version"],
    "Java": ET.parse(root / "sdk/java/pom.xml").getroot().findtext("{http://maven.apache.org/POM/4.0.0}version"),
}
for sdk, actual in versions.items():
    if actual != version:
        raise SystemExit(f"{sdk} SDK version {actual!r} does not match release {version!r}")
PY

mkdir -p dist/release

npm ci --prefix sdk/typescript
npm run build --prefix sdk/typescript
(
  cd sdk/typescript
  npm pack --pack-destination "$repo_root/dist/release" \
    --cache "$repo_root/dist/npm-cache"
)

git archive --format=tar.gz --prefix=semantic-validator/ \
  -o "dist/release/semantic-validator-python-${version}.tar.gz" \
  HEAD LICENSE sdk/python
git archive --format=tar.gz --prefix=semantic-validator/ \
  -o "dist/release/semantic-validator-go-${version}.tar.gz" \
  HEAD LICENSE go.mod sdk/go
git archive --format=tar.gz --prefix=semantic-validator/ \
  -o "dist/release/semantic-validator-java-${version}.tar.gz" \
  HEAD LICENSE sdk/java

(
  cd dist/release
  sha256sum \
    "semantic-validator-sdk-${version}.tgz" \
    "semantic-validator-python-${version}.tar.gz" \
    "semantic-validator-go-${version}.tar.gz" \
    "semantic-validator-java-${version}.tar.gz" > SHA256SUMS.txt
  sha256sum -c SHA256SUMS.txt
)

echo "Release assets ready in dist/release/"
