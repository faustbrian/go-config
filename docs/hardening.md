# Hardening evidence and findings

## Scope and result

The production code is covered at exactly 100% statement coverage. Tests prove
atomic plan loading, stable precedence, presence states, merge truth tables,
strict decode, discovery containment, parser bounds, provenance, redaction,
validation ordering, cancellation, immutable copies, and filesystem lifecycle
behavior. Race tests cover parallel repeated loads. Six fuzz targets cover
structured parsers, dotenv/interpolation, environment mapping, dynamic decode
tags/destination types, filesystem boundaries, and discovery policies.

No high or medium release finding remains in the implemented core scope. The
native Infisical adapter is not implemented and therefore has no core SDK or
credential lifecycle to audit. It remains an optional post-core deliverable.

## Threat and failure matrix

| Area | Failure or threat | Enforced response |
|---|---|---|
| source | missing optional | suppress only wrapped `ErrNotFound` |
| source | unreadable/malformed | fail complete candidate |
| merge | structural type change | typed conflict, no snapshot |
| decode | unknown/required/overflow | aggregated safe field errors |
| validation | error or panic | safe deterministic failure, no snapshot |
| parser | byte/depth/key growth | explicit bounded error |
| dotenv | cycle/expansion growth | bounded interpolation error |
| discovery | traversal/symlink escape | reject outside lexical/resolved root |
| filesystem | partial read/cancel/close | fail complete load |
| custom source | cycle/depth/key growth | bounded canonicalization failure |
| typed schema | recursive/deep structs | constructor-time schema failure |
| secret | supported format/marshal | `[REDACTED]` |
| snapshot | retained mutable value | deep defensive copy |

## Intentional limitations

- No automatic hot reload, generation publication, or consumer rotation in v1.
- No physical secret-memory erasure guarantee in Go.
- No cryptographic file authenticity verification.
- No executable configuration or arbitrary expression language.
- No remote source or native Infisical adapter in core.
- Cross-platform permission and symlink behavior remains constrained by host OS
  facilities; CI runs Linux, macOS, and Windows matrices.

## Verification commands

`make check` runs formatting, API compatibility, unsafe/cgo checks, vet, race,
exact coverage, all fuzz smoke targets, benchmark smoke, docs/examples, and
reachable vulnerability scanning. GitHub additionally runs golangci-lint,
dependency review, scheduled fuzzing/benchmarks, and tagged release validation.

Performance evidence and operating budgets are documented in
[operations.md](operations.md). The security model is in
[security.md](security.md).
