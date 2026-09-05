# Troubleshooting

## Construction fails

Use `errors.Is` to distinguish `ErrClientRequired` from `ErrInvalidOptions`.
For invalid options, verify that `Name` and `SecretID` are non-empty, that
`SecretID` is no longer than 2,048 bytes, and that `MaximumBytes` is between 0
and 65,536. If an ARN or generated identifier exceeds the bound, correct the
configured identifier rather than truncating it, because truncation can select
a different secret. Constructing the source performs no provider I/O, so
correct the configuration before retrying startup.

## A load reports that the source was not found

Check `errors.Is(err, config.ErrNotFound)`. Confirm the secret identifier,
region, and optional version ID or stage using a trusted administrative path.
If the source is intentionally optional, let the enclosing `config.Plan`
apply its documented optional-source behavior. Do not log secret payloads while
diagnosing selection.

## A load returns `ErrOperation`

The formatted error intentionally redacts provider details. Within a trusted
diagnostic boundary, use `errors.As` against AWS SDK error types to distinguish
credential, IAM, KMS, region, endpoint, throttling, and availability failures.
Correct workload identity, least-privilege access, region, or client retry
configuration as appropriate. Do not print the wrapped provider error or AWS
response because either may contain sensitive operational data.

## A response is rejected

`ErrInvalidResponse` means the client returned no payload, both string and
binary payloads, an empty payload, or a payload above the configured bound.
Other parse errors indicate that the selected value is not one valid JSON
object. Inspect only metadata such as payload form and byte count, then replace
the secret with bounded valid JSON; do not dump the value into logs.

## A load is cancelled or times out

Check `errors.Is` against `context.Canceled` and `context.DeadlineExceeded`.
When cancellation is already observable, the adapter does not invoke the
client. Otherwise it passes the same context to one client invocation, whose
network and retry behavior belongs to the client. Retry only under an
application-owned startup or refresh policy, preferably with an explicit
version ID when the selected value must remain stable.

## A rotated value is not visible

The adapter has no cache or refresh loop. Applications that load once at
startup continue using their existing decoded configuration. Perform another
explicit load, validate the complete replacement, and define transition and
rollback behavior in the application before swapping runtime configuration.
