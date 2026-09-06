# Compatibility Policy

Each releasable directory is an independent Go module and follows semantic
versioning. Root-module releases use `v<version>` tags. The separately
releasable AWS Secrets Manager adapter uses
`adapters/awssecretsmanager/v<version>` tags. A directory prefix is never added
to the root module's tag.

The root and AWS Secrets Manager modules are stable v1 libraries. Their minimum
supported Go version is 1.26.6, and repository verification currently tests
exactly Go 1.26.6. The adapter supports AWS Secrets Manager through the AWS SDK
for Go v2 and delegates region, endpoint, credentials, transport, and retry
compatibility to the caller-provided client.

Before `v1`, minor releases MAY contain reviewed breaking changes, but every
break MUST be documented with migration guidance. Patch releases MUST remain
backward compatible. At and after `v1`, incompatible exported API or documented
behavior changes require a new major version.

Compatibility includes exported Go APIs, error classification, serialization,
protocol behavior, persistence schemas, environment variables, command output,
resource ownership, ordering, retry/idempotency semantics, and documented
defaults. A compile-compatible change can still be behaviorally breaking.

Specification-backed modules MUST NOT diverge from their declared standards.
Ambiguities require documented decisions and stable tests. Deprecated APIs
follow [`DEPRECATION.md`](DEPRECATION.md).

Observable structured-format interpretations follow the
[specification decision register](docs/specification-decisions.md) and require
compatibility review.
