# config

[![CI](https://github.com/faustbrian/go-config/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/faustbrian/go-config/actions/workflows/ci.yml)
[![CodeQL](https://img.shields.io/badge/CodeQL-required-blue)](https://github.com/faustbrian/go-config/actions/workflows/ci.yml)
[![Mutation](https://img.shields.io/badge/mutation-local_or_scheduled-blue)](CONTRIBUTING.md#verification)
[![Documentation](https://img.shields.io/badge/docs-structural_checks_in_CI-blue)](docs/)
[![Go Reference](https://pkg.go.dev/badge/github.com/faustbrian/go-config/v3.svg)](https://pkg.go.dev/github.com/faustbrian/go-config/v3)
[![Release](https://img.shields.io/github/v/release/faustbrian/go-config?sort=semver)](https://github.com/faustbrian/go-config/releases)
[![Go](https://img.shields.io/badge/go-1.27.0-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`config` loads explicit, layered configuration sources into caller-owned Go
structs. It provides deterministic precedence, strict decoding, defensively
copied snapshots with typed-shape limitations, declared provenance, validation
orchestration and redacted secret formatting without introducing global state or
implicit filesystem discovery.

Main tracks v3 of the root library. It requires Go 1.27.0; published v1
releases retain their original Go support contract. Use published tags rather than untagged main. See
the [migration guide](docs/migration.md) before adopting the new identities.

## Install

```console
go get github.com/faustbrian/go-config/v3@v3.0.0
```

Use the independently released AWS Secrets Manager adapter only when an
application reads configuration directly from AWS Secrets Manager. Root-v3
applications use its v3 line; its v2 line builds on root v2. See the
[migration guide](docs/migration.md#to-v3).

## Five-minute quickstart

```go
package main

import (
	"context"
	"fmt"

	config "github.com/faustbrian/go-config/v3"
	"github.com/faustbrian/go-config/v3/defaults"
	"github.com/faustbrian/go-config/v3/environment"
	jsonsource "github.com/faustbrian/go-config/v3/json"
	"github.com/faustbrian/go-config/v3/programmatic"
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

The v3 module removes the former `configservice` facade path; see the
[migration guide](docs/migration.md#to-v3).

## What is included

Strict JSON, YAML, TOML, dotenv, environment, map, byte, reader, `fs.FS`, and
explicit-file sources compose with bounded discovery, merging, typed values,
validation, copied snapshots, provenance, and an optional service adapter.

The independently versioned
[`github.com/faustbrian/go-config/adapters/awssecretsmanager/v3`](adapters/awssecretsmanager/README.md)
module adapts one bounded AWS Secrets Manager JSON document into the same
`config.Source` contract. Use it only when an operator, CSI driver, or sidecar
does not already materialize the secret as an environment variable or file.

## Documentation

For shared package selection, ownership, construction, and lifecycle guidance,
see the versioned [Golib ecosystem index](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/README.md)
and its [Foundations family](https://github.com/faustbrian/go-library-tools/blob/v1.8.4/docs/ecosystem/design-language.md#package-families-and-selection).

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
they open during a load; returned snapshots are values rather than live resource
handles, subject to [typed copying limitations](docs/api.md). The AWS adapter
follows the same no-shutdown boundary and leaves its client, transport,
credentials, retries, and refresh scheduling with the caller.

## Development

Use Go 1.27.2 for development; the public minimum remains Go 1.27.0.
`make check` and `make ci` run bounded local checks with the public pinned
verifier. Use `make check-full` for full assurance and online specification
checks. See [CONTRIBUTING.md](CONTRIBUTING.md) for standalone Go tests,
focused fuzzing, optional mutation and selected integration/release qualification.

## License

MIT. See [LICENSE](LICENSE).
