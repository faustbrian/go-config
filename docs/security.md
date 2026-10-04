# Security and threat model

## Model version and scope

This is threat model `CONFIG-TM-1.1`, reviewed 2026-10-05. It applies to the
root module's current unreleased v1 line. Reassess it when discovery,
filesystem loading, source trust, platform support, or secret-handling
boundaries change.

## Protected properties

The package protects atomic publication, deterministic precedence, root-bounded
discovery, bounded parsing/interpolation, immutable returned snapshots, and
secret-safe supported diagnostics. Inputs are considered untrusted. Callers are
responsible for selecting trustworthy roots, filesystems, environment sources,
validators, and custom decoding hooks.

`Secret` formats and marshals as `[REDACTED]`. Source sensitivity and
`config:",secret"` metadata flow into provenance without storing the field
value. Decode, mapping, default, interpolation, validation, panic, and parser
errors describe paths/categories only. Never include secret values in source
names, environment names, or file paths. Arbitrary decoder, validator, parser,
environment, and default causes are replaced by safe wrappers throughout the
public error chain while retaining `errors.Is` sentinel checks. Concrete custom
cause recovery through `errors.As` is intentionally not supported.
Cause-bearing errors also override detailed formatting and text/JSON marshaling,
so structured loggers cannot reflect through exported cause fields. Discovery
policy errors omit rejected paths and arbitrary platform error text.

## Limits and filesystem policy

Every parser and environment source has explicit resource limits. Discovery has
candidate, result, and upward-depth limits. Dotenv expansion has depth and byte
limits. Context cancellation is checked during reads and recursive processing.
Set tighter limits for known small configurations.

For remote or adversarial filesystem implementations, use `ContextFS`,
`ContextFile`, and `ContextCloser` so open/read/stat/close calls observe
deadlines, and `GenerationFile` so same-size rewrites have an explicit stable
identity. Ordinary `fs.FS` implementations are a caller-owned liveness and
metadata trust boundary. Reads fail after repeated zero-byte progress instead
of spinning indefinitely.
Context-aware close receives a fresh one-second cleanup context so a canceled
load cannot suppress resource release.

Use a discovery root, reject symlinks unless needed, and enable `OwnerOnly` for
secret-bearing local files. Kubernetes projected-volume modes and ownership are
platform settings; verify them in the pod security context. Optional sources
only suppress `ErrNotFound`, never permission, syntax, decoding, or validation
errors.

`filesystem.FromDiscovered` binds the approved directory identity at source
construction. Each load uses `os.Root` confinement and compares the regular
file identity before, during, and after open. A swapped final symlink is
rejected, an outside-root symlink cannot be followed on platforms where
`os.Root` provides its documented confinement, and a concurrent identity
change returns `ErrSourceChanged`. `filesystem.FromPath` intentionally remains
path-following for caller-approved hot-reload paths.

## Versioned risk register

| Risk | Status in CONFIG-TM-1.1 | Owner | Rationale | Control or required mitigation | Review condition |
|---|---|---|---|---|---|
| CONFIG-FS-001: a discovered target or its approved directory is replaced before load to escape the trusted root | Mitigated on operating systems with descriptor-backed `os.Root` confinement | Library maintainers | The discovered path is attacker-mutable after discovery, so lexical validation alone cannot authorize the later open. | Directory identity binding, root-confined open, final-symlink rejection, and pre/open/post file identity comparison. | Reassess when discovery, filesystem loading, or the minimum Go version changes. |
| CONFIG-FS-002: an attacker writes a stable regular file inside the approved directory | Accepted caller boundary | Deploying application owner | A generic configuration loader cannot distinguish authorized bytes from malicious bytes at a caller-approved path. | Read-only ownership, deployment policy, and artifact authenticity controls; this library does not authenticate local file contents. | Reassess if the library adds content-authenticity policy or accepts a new source trust model. |
| CONFIG-FS-003: JavaScript or Plan 9 directory replacement, or OS-level bind-mount/device namespace changes | Accepted platform boundary | Library maintainers | The supported standard-library confinement primitive cannot provide stronger guarantees on these platforms or against namespace administration. | Do not treat attacker-writable roots as confined on JavaScript; avoid hostile directory renames on Plan 9; protect mount and namespace administration on all platforms. | Reassess when `os.Root` platform guarantees change or supported platforms change. |

## Known limits

Go cannot guarantee physical zeroization of strings, copied bytes, stack values,
or garbage-collected objects. `Secret` reduces accidental disclosure; it is not
a protected-memory container. `Reveal` returns plaintext intentionally. A
caller can still leak it through logging, metrics, tracing, panic values,
validators, or downstream libraries.

The intermediate source tree accepts only canonical scalar, object, array,
null, and deletion-marker values. Typed custom values with private maps,
slices, pointers, interfaces, functions, or channels are rejected unless they
are library-owned clone-aware wrappers. `time.Time` is treated as a standard
immutable scalar. Keep custom scalar private state immutable.

Custom source trees are cycle-checked, cancellation-aware, and bounded to 64
levels, 100,000 aggregate keys, and 100,000 aggregate array elements before
merge. Defaults and environment schema walkers reject recursive types and
nesting beyond 64 structs. Arbitrary source errors are redacted at the plan
boundary even when the source is not marked sensitive; use `errors.Is` for
stable identity checks.

Context-aware decode hooks receive already-canonical bounded input and the load
context. Legacy `encoding.TextUnmarshaler` and `ValueUnmarshaler` implementations
remain trusted synchronous application code. Go cannot safely preempt an
arbitrary blocking hook without risking a leaked goroutine, so untrusted hooks
must implement the context-aware interfaces and cooperate with cancellation.

No cryptographic authenticity check is performed on local configuration files.
Use trusted deployment mechanisms, read-only mounts, image signing, Kubernetes
RBAC, and secret-store policy. The library does not rotate, create, update, or
delete secrets.

## Infisical boundary

Kubernetes should normally use Infisical Operator, CSI, or Agent delivery. Core
then consumes environment variables or mounted files and has no Infisical SDK,
credential, retry, or token-refresh lifecycle. A future native adapter must be a
separate module, read-only, explicitly scoped, bounded, fail-closed for required
secrets, and independently audited. It is not implemented in core.

Report suspected vulnerabilities according to [SECURITY.md](../SECURITY.md).
