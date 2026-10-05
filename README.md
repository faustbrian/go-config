# config

[![CI](https://github.com/faustbrian/go-config/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-config/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-config/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Mutation](https://img.shields.io/badge/mutation-100%25_required-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-checked_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-config/v2.svg)](https://pkg.go.dev/github.com/faustbrian/go-config/v2)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-config?sort=semver)](https://github.com/faustbrian/go-config/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`config` loads explicit, layered configuration sources into caller-owned Go
structs. It provides deterministic precedence, strict decoding, immutable
snapshots, safe provenance, validation orchestration, and redacted secrets
without introducing global state or implicit filesystem discovery.

Main tracks v2 of the root library and its independently released AWS Secrets
Manager adapter. Both require Go 1.27.0; published v1 releases retain their
original Go support contract. Use published tags rather than untagged main. See
the [migration guide](docs/migration.md) before adopting the new identities.

## Install

```console
go get github.com/faustbrian/go-config/v2@v2
go get github.com/faustbrian/go-config/adapters/awssecretsmanager/v2@v2
```

The root v2 tag is public. Use the second command after the adapter's own v2
tag is published, and only when an application reads configuration directly
from AWS Secrets Manager. Its v2 source on main already consumes public root v2.

## Five-minute quickstart

```go
package main

import (
	"context"
	"fmt"

	config "github.com/faustbrian/go-config/v2"
	"github.com/faustbrian/go-config/v2/defaults"
	"github.com/faustbrian/go-config/v2/environment"
	jsonsource "github.com/faustbrian/go-config/v2/json"
	"github.com/faustbrian/go-config/v2/programmatic"
)

type Settings struct {
	Host  string        `config:"host" env:"HOST" default:"127.0.0.1"`
	Port  int           `config:"port" env:"PORT" default:"8080"`
	Token config.Secret `config:"token,secret" env:"TOKEN"`
}

func main() {
	forDefaults := defaults.For[Settings]
	base, _ := forDefaults("defaults")
	file, _ := jsonsource.Bytes(
		[]byte(`{"host":"service.internal","port":9000}`),
		jsonsource.Options{Name: "config.json"},
	)
	env, _ := environment.EnvironFor[Settings](
		[]string{"PORT=9443", "TOKEN=not-printed"},
		environment.Options{Name: "environment"},
	)
	override, _ := programmatic.Overrides(
		"command-line", map[string]any{"host": "localhost"},
	)

	plan, _ := config.NewDefaultPlan(config.DefaultSources{
		Defaults: []config.Source{base},
		DiscoveredBase: []config.Source{file},
		Environment: []config.Source{env},
		Overrides: []config.Source{override},
	})
	load := config.Load[Settings]
	snapshot, err := load(context.Background(), plan)
	if err != nil {
		panic(err)
	}

	settings := snapshot.Value()
	fmt.Printf("%s:%d token=%s\n", settings.Host, settings.Port, settings.Token)
	// localhost:9443 token=[REDACTED]
}
```

All constructors return errors; production code should handle them. The
omitted checks above keep the precedence example compact. A complete version is
in [examples/quickstart](examples/quickstart/main.go).

## Service command loading

The target-oriented [`adapters/service`](adapters/service) package adapts a
typed plan to `service.CommandSpec.Load`. Import it conventionally as
`configservice`. It loads only after command selection and before component
construction. Local dotenv files require the explicit `Local` option; process
environment and caller overrides retain the standard precedence.

```go
loader, err := configservice.New(configservice.Options[Settings]{
    Local: true,
    Dotenv: &configservice.Dotenv{
        FS: os.DirFS("."),
        Path: ".env",
        Options: dotenv.Options{Name: "local-dotenv", Prefix: "APP_"},
    },
    Environment: &environment.Options{
        Name: "process-environment",
        Prefix: "APP_",
    },
})
if err != nil {
    return err
}

command := service.CommandFor(service.CommandSpec[Settings]{
    Name: "serve",
    Kind: service.CommandKindLongRunning,
    Load: loader,
    Build: build,
})
```

The adapter owns no resource, performs no retries, and does not reload
configuration. The caller owns every source and decides whether repeated loads
are safe.

Within v2, `github.com/faustbrian/go-config/v2/configservice` remains an
API-compatible facade for the target-oriented v2 service adapter. V1 consumers
must first migrate their imports; choosing between the two v2 adapter paths
does not require a coordinated migration.

## What is included

Strict JSON, YAML, TOML, dotenv, environment, map, byte, reader, `fs.FS`, and
explicit-file sources compose with bounded discovery, merging, typed values,
validation, immutable snapshots, provenance, and an optional service adapter.

The independently versioned
[`github.com/faustbrian/go-config/adapters/awssecretsmanager/v2`](adapters/awssecretsmanager/README.md)
module adapts one bounded AWS Secrets Manager JSON document into the same
`config.Source` contract. Use it only when an operator, CSI driver, or sidecar
does not already materialize the secret as an environment variable or file.

## Documentation

For shared package selection, ownership, construction, and lifecycle guidance,
see the versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/README.md)
and its [Foundations family](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md#package-families-and-selection).

Use the [documentation index](docs/README.md) for formats, layering,
discovery, Kubernetes, security, migration, and package-author guidance.
Observable structured-format choices are recorded in the
[specification decision register](docs/specification-decisions.md).

## Design boundaries

The package does not load automatically, mutate process environment, traverse
parents by default, execute configuration code, hot-reload snapshots, manage
secrets, or define vendor credential structs. Kubernetes applications should
normally receive Infisical values through the Operator, CSI, or Agent and load
the resulting environment variables or files normally.

The root module starts no background goroutines, watchers, refresh loops, or
retry loops and exposes no `Close` or `Shutdown` method. Callers own supplied
sources and collaborators, decide when loads occur, and finish in-flight loads
before releasing those collaborators. Built-in sources close the transient I/O
they open during a load; returned snapshots remain immutable values rather than
live resource handles. The AWS adapter follows the same no-shutdown boundary
and leaves its client, transport, credentials, retries, and refresh scheduling
with the caller.

## Development

Run `make check`. See [CONTRIBUTING.md](CONTRIBUTING.md) for focused and release
verification.

## License

MIT. See [LICENSE](LICENSE).
