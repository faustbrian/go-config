# Security

[Documentation index](README.md)

Use workload identity and the AWS SDK default credential chain. Do not provide
static AWS access keys to this adapter. IAM should allow only
`secretsmanager:GetSecretValue` for the configured secret and the KMS decrypt
permission required by that secret.

The source identifier and version selectors are configuration metadata, not
secret material. Payload values are marked sensitive and provider error details
are redacted from formatting. The transient adapter-owned payload copy is
cleared after parsing, but SDK responses, strings and parser copies are not
securely erased. Size admission follows SDK materialization and the local copy,
not a whole-HTTP-memory budget; the returned configuration document necessarily
owns its decoded values for the lifetime chosen by the caller.

Because the source is sensitive, a root snapshot from go-config v3.1.0 or later
records no `SourceDigest` for it. An unsalted digest of a short secret payload
would let anyone holding it test guesses offline.

Avoid logging source documents, typed settings, AWS responses, or wrapped
provider causes. Rotation and refresh must be an explicit application workflow
with validated transition and rollback behavior.

Continue with [API](api.md).
