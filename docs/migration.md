# Migration guide

[Documentation index](README.md)

## To v2 and Go 1.27

The v2 release line is maintained on main. Its new module identities make
the Go 1.27 minimum explicit without changing the Go 1.26.6 contract of already
published v1 releases. Upgrade the toolchain before adopting the new tags.

Replace root imports with `github.com/faustbrian/go-config/v2`, retaining each
subpackage suffix after `/v2`. The optional AWS adapter instead uses its own
identity, `github.com/faustbrian/go-config/adapters/awssecretsmanager/v2`.
Its sources return root-v2 types; do not mix them with root-v1 plans or loaders.
Root v2.0.0 and the independent AWS adapter v2.0.1 are publicly available.
Select each published module version explicitly when using the adapter.

Root tags are `v2.0.0`; adapter tags are
`adapters/awssecretsmanager/v2.0.0`. Both are created from main without separate
version directories or branches. The existing v1 tags remain unchanged.

## To root v2.0.1

Main prepares a compatible root patch; after release qualification and public
availability, require `github.com/faustbrian/go-config/v2@v2.0.1`. Keep the same
imports and Go 1.27.0 minimum. No new AWS adapter release is necessary: an
application using published AWS adapter v2.0.1 can add the corrected root
requirement explicitly, without changing the adapter's own dependency pin.

The patch restores snapshot isolation through array references and validates
Optional payloads before cloning or callbacks. Handle `SnapshotValueError`
from unsupported or cyclic custom decoded state and `TreeCycleError` from
cyclic programmatic containers. Finite acyclic inputs remain accepted; no new
constructor depth or count cap is introduced. Deletion through canonical
Sources consumes nested object-key tombstones; programmatic marker-struct
rejection remains unchanged. A context canceled by the final validator now
returns its context error rather than publishing a candidate.

## To the target-oriented service adapter

New integrations should import
`github.com/faustbrian/go-config/v2/adapters/service`, conventionally aliased as
`configservice`. Its generic `Options[T]` and `Loader[T]` shapes,
`ErrInvalidOptions`, `OptionsError`, `Dotenv`, and `New[T]` behavior match the
v2 facade path. Types remain named by their import paths, so migrate an import
as one source change rather than mixing both packages in one assignment.

After migrating from v1, consumers may use
`github.com/faustbrian/go-config/v2/configservice` as a compatibility facade.
Switching between this facade and the target-oriented v2 adapter does not
require a coordinated source migration; adopting v2 still requires the module
and import changes described above.

## From direct `os.Getenv`

Define one root struct at the application composition root. Add `config` tags
for document names and explicit `env` tags for environment names. Replace
scattered reads with one `environment.ProcessFor[T]`, one plan, and one boot
load. Pass the resulting typed value or owned component sub-structs to
constructors.

Before:

```go
port, _ := strconv.Atoi(os.Getenv("PORT"))
token := os.Getenv("TOKEN")
```

After:

```go
type Settings struct {
	Port  int           `config:"port,required" env:"PORT"`
	Token config.Secret `config:"token,required,secret" env:"TOKEN"`
}

source, err := environment.ProcessFor[Settings](environment.Options{
	Name: "environment",
})
```

Handle constructor and load errors at startup. Use explicit test environment
slices instead of `t.Setenv` where possible. Do not retain a global mutable
configuration singleton.

## From Laravel configuration

Laravel commonly combines PHP config files, `.env`, `env()`, service-container
lookups, and runtime `config()` mutation. Translate the final public contract,
not executable PHP behavior:

1. Define caller-owned Go structs for application settings and package-owned
   sub-structs for reusable integrations.
2. Move static PHP defaults to typed `default` tags or a programmatic defaults
   source.
3. Translate deployable config files to strict JSON, YAML, or TOML.
4. Map `.env` only when explicitly requested; keep process environment above it
   in `NewDefaultPlan`.
5. Replace `config('path')` reads with typed field access.
6. Replace `config([...])` runtime writes with explicit immutable overrides
   before loading or with application state outside configuration.
7. Replace service-container resolution with ordinary constructor arguments.
8. Add validators for cross-field rules that Laravel config previously assumed.

PHP config files can execute code; this library deliberately cannot. Compute
dynamic values in the composition root and pass them through a programmatic
source. Preserve secret ownership in deployment tooling rather than committing
plaintext translations.

During migration, compare a sanitized field inventory and provenance, never a
dump containing values. Roll out with the old and new loaders reading the same
inputs only if the old path can be observed safely; publish one authoritative
configuration to consumers.

Continue with [API reference](api.md).
