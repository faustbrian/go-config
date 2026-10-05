# Compatibility Policy

Each releasable directory is an independent Go module and follows semantic
versioning. Root-module releases use `v<version>` tags. The separately
releasable AWS Secrets Manager adapter uses
`adapters/awssecretsmanager/v<version>` tags. A directory prefix is never added
to the root module's tag.

Main prepares v2 of the root and AWS Secrets Manager modules. Both require
Go 1.27.0, which repository verification tests exactly. Published v1 releases
retain their original Go 1.26.6 support contract; the v2 identities and tags
must be published before applications can adopt them. The adapter supports
AWS Secrets Manager through the AWS SDK for Go v2 and delegates region,
endpoint, credentials, transport, and retry compatibility to the caller's client.

Before `v1`, minor releases MAY contain reviewed breaking changes, but every
break MUST be documented with migration guidance. Patch releases MUST remain
backward compatible. At and after `v1`, incompatible exported API or documented
behavior changes require a new major version.

Compatibility includes exported Go APIs, error classification, serialization,
protocol behavior, persistence schemas, environment variables, command output,
resource ownership, ordering, retry/idempotency semantics, and documented
defaults. A compile-compatible change can still be behaviorally breaking.

The target-oriented `adapters/service` package owns service command loading.
Within v2, `configservice` remains an API-compatible facade
with its existing generic option and loader definitions, shared sentinel
identity, equivalent structured errors, source precedence, cancellation, and
caller-owned lifecycle semantics. V1 consumers must migrate their module and
import identities before using either v2 package; facade compatibility does
not make nominal v1 and v2 types interchangeable.

Specification-backed modules MUST NOT diverge from their declared standards.
Ambiguities require documented decisions and stable tests. Deprecated APIs
follow [`DEPRECATION.md`](DEPRECATION.md).

Observable structured-format interpretations follow the
[specification decision register](docs/specification-decisions.md) and require
compatibility review.
