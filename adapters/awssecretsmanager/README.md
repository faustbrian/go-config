# AWS Secrets Manager configuration source

This stable adapter loads one bounded JSON configuration document from AWS
Secrets Manager into a sensitive
[`github.com/faustbrian/go-config`](https://pkg.go.dev/github.com/faustbrian/go-config)
source. It leaves AWS credential resolution, retries, IAM, secret creation,
rotation, and refresh scheduling to the caller.

It requires Go 1.26.6 or newer. The adapter is released independently from the
root module under `adapters/awssecretsmanager/v*` tags.

## Install

```console
go get github.com/faustbrian/go-config/adapters/awssecretsmanager@latest
```

Import the canonical module path directly:

```go
import "github.com/faustbrian/go-config/adapters/awssecretsmanager"
```

## Quick start

```go
awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
if err != nil {
    return err
}
source, err := awssecretsmanager.New(
    secretsmanager.NewFromConfig(awsConfig),
    awssecretsmanager.Options{
        Name:     "runtime-secrets",
        SecretID: "track/production/runtime",
    },
)
if err != nil {
    return err
}
document, err := source.Load(ctx)
if err != nil {
    return err
}
_ = document
```

The source reads `AWSCURRENT` by default. Supply `VersionID`, `VersionStage`,
or both when startup must be pinned to an immutable version. Place the source
below process-environment overrides in an explicit `config.Plan`. See the
[executable package example](example_test.go) for a self-contained client.

## Package map

| Package | Use |
| --- | --- |
| `github.com/faustbrian/go-config/adapters/awssecretsmanager` | Construct one explicit AWS Secrets Manager-backed `config.Source`. |

This module has no public subpackages. It is an optional adapter for the
independently released root `github.com/faustbrian/go-config` module.

## Construction and lifecycle

`New(Client, Options)` validates configuration without network I/O and returns
a stateless `config.Source`. The caller owns the AWS SDK configuration, client,
credentials, retry policy, and client transport for their full lifetimes.
`Load(ctx)` observes caller cancellation, performs one `GetSecretValue` call,
and returns after parsing one JSON object; it starts no goroutines and retains
no response or payload buffer.

The source is safe for concurrent use when the caller-provided `Client` is safe
for concurrent use. The adapter has no `Close` or `Shutdown` method because it
owns no persistent resource or background work. The caller closes any transport
or other resource owned by its client.

## Guarantees

- exactly one explicit secret identifier is read per load;
- payload work is bounded by the AWS 65,536-byte service limit;
- both AWS string and binary JSON values are supported;
- missing secrets map to `config.ErrNotFound` for optional-source semantics;
- provider error details and all loaded fields are marked sensitive; and
- no process environment, global AWS configuration, cache, or goroutine is
  owned by the adapter.

## Tradeoffs

Each load performs one provider read. Callers should load once during process
startup unless they intentionally own refresh, version transition, and failure
semantics. The adapter accepts JSON objects only and does not flatten dotenv
text or mutate `os.Environ`.

Use the root module's environment or filesystem sources when an operator, CSI
driver, or sidecar already materializes the secret. Do not use this adapter as
a credential provider, cache, rotation controller, or one-secret-per-field
loader.

## Errors and security

Construction rejects a missing client or invalid options before provider I/O.
`Load` preserves cancellation and deadline errors, maps AWS missing-secret
responses to `config.ErrNotFound`, and exposes stable adapter categories through
`errors.Is`. Provider causes remain available through `errors.Is` and
`errors.As`, but formatted adapter errors redact provider details.

The adapter marks the returned document sensitive, clears its transient payload
copy after parsing, and never logs secrets. Callers remain responsible for
least-privilege IAM and KMS policy, safe handling of decoded values, and the
observability and shutdown lifecycle of their AWS client. See the full
[security guidance](docs/security.md) and report vulnerabilities through the
[repository security policy](../../SECURITY.md).

## Compatibility and operations

The supported backend is AWS Secrets Manager through AWS SDK for Go v2. String
and binary secret values must contain one JSON object. Each load performs one
provider request, so latency, availability, retries, and cost follow the
caller-configured AWS client. See [compatibility](docs/compatibility.md),
[architecture and performance](docs/architecture.md), and
[adoption guidance](docs/adoption.md) before adding refresh behavior.

## Documentation

For shared package selection, ownership, construction, and lifecycle guidance,
see the versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Foundations family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

- [API](docs/api.md)
- [Adoption](docs/adoption.md)
- [Architecture](docs/architecture.md)
- [Compatibility](docs/compatibility.md)
- [Security](docs/security.md)
- [FAQ](docs/faq.md)
- [Troubleshooting](docs/faq.md)
- [Changelog](CHANGELOG.md)
- [Support](../../SUPPORT.md)
- [Contributing](../../CONTRIBUTING.md)
- [License](../../LICENSE)

## Development

```console
make check
```

## Testing helpers

The module exports no test helper. Implement the narrow `Client` interface in
tests, as demonstrated by the [executable example](example_test.go), to control
provider responses without global AWS state.

## License

[MIT](../../LICENSE).

## Related packages

See the root [documentation index](../../docs/README.md) for source ownership,
security, and operations guidance.
