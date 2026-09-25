# Open-source project

[Español](OPEN_SOURCE.es.md) · [Project README](../README.md)

The [GitHub repository](https://github.com/eduardoArequipa/semantic-validator)
contains the Go server, four SDKs, documentation, and beta examples. The source
and downloadable packages use [Apache License 2.0](../LICENSE). Semantic
Validator is free and open source; there is no paid hosted plan. Jev is a
separate service used with each developer's own key and account.

## Before publishing a release or package

1. Confirm code ownership and review third-party dependency licenses and
   notices before publishing source or packages.
2. The previously shared Jev token was revoked and removed from the VPS
   `.env`. Review every other local copy or history before publishing. Do not
   add credentials to Git.
3. Run `go test ./...`, `go vet ./...`, SDK tests, and
   `python3 tools/evaluate.py --dry-run`. The proposed GitHub Actions workflow
   needs no live key.
4. Review the exact files to be added to Git. `.env`, key data, private
   evaluation cases, binaries, and generated build files are excluded by
   `.gitignore`. Four SDK archives in `internal/api/docs/downloads/` are
   served by the website; inspect them before publishing. The old TypeScript
   demo archive remains in the source tree but is not embedded or served.
5. The public Go module path is
   `github.com/eduardoArequipa/semantic-validator`. The TypeScript, Python,
   and Java SDKs are not yet published to package registries. Review Jev/TypeSafe
   usage and branding terms separately from this project's
   source-code license. The public [TypeSafe agreement](https://typesafe.ai/legal/mca)
   prohibits making its service available as a standalone service. Seek written
   permission for any third-party hosted proxy; BYOK self-hosting is the
   recommended open-source path, subject to each operator's own terms.

## Hosted API paused

The hosted API is in `DOCS_ONLY` mode: the website and documentation remain
available, while `/v1/*` returns HTTP 503 without calling Jev. Its former
individual-key data remains on the VPS for recovery, but the revoked Jev token
has been removed from the VPS `.env`. Do not broaden access or market a
standalone Jev proxy without written permission from TypeSafe. There is no
signup, billing, or paid hosted offering.
