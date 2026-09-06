# Security policy

## Supported versions

The root module and the separately releasable AWS Secrets Manager adapter each
have a published v1 line. The latest v1 patch release for each module is
supported unless announced otherwise. Fixes land on the default branch before
the affected module is released independently.

## Reporting a vulnerability

Use this repository's
[private security advisory](https://github.com/faustbrian/go-config/security/advisories/new)
form. Do not open a public issue containing secrets, credentials, exploit
details, or vulnerable deployment information. Include the affected module and
version, configuration source, minimal reproduction, impact, and any known
mitigations.

## Scope

Secret disclosure, path traversal, symlink/root escape, parser resource
exhaustion, partial snapshot publication, unsafe optional-source suppression,
and validation or interpolation bypasses are security relevant. The package
does not claim physical memory zeroization: Go strings and garbage collection
cannot provide that guarantee.

See the full [security model](docs/security.md).
