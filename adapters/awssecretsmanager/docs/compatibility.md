# Compatibility

The adapter supports AWS Secrets Manager string and binary secret values that
contain one JSON object. It uses the AWS SDK v2 `GetSecretValue` contract and
the current `github.com/faustbrian/go-config` document model. The module
requires Go 1.26.6 or newer, matching its `go.mod` directive and repository
support policy. Its supported platform classification is `portable-go`: the
adapter contains no OS-specific code and requires no local service process.

Adding optional `Options` fields is compatible. Changing the default version
stage, source priority, JSON object requirement, missing-secret mapping, or
error redaction is behaviorally breaking and requires a major release.
