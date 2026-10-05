# Security policy

## Supported versions

The root module and the separately releasable AWS Secrets Manager adapter each
have a published v1 line. Reports for the latest public v1 releases remain
supported during the v2 transition. Main tracks the Go 1.27-only v2 lines;
untagged main is not a supported public release. Published root v2.0.0 includes
the discovery identity fix; v1.1.0 does not include it. The adapter's own v2
release remains independent. Fixes
land on main before the affected module is released independently; no v1
backport is promised by the v2 plan.

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
