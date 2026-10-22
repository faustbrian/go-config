# Security policy

## Supported versions

Main tracks root v3; untagged main is not a supported public release. Reports
against the latest published root v3 and v2 patches and the latest AWS
Secrets Manager adapter v2 patch are supported. The latest public v1 releases
receive security fixes through at least 2026-12-31; published root v2.0.0
includes the discovery identity fix that v1.1.0 lacks. Fixes land on main before the affected module is released
independently.

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
