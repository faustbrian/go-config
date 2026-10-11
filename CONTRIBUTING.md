# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected module's docs.
2. Run ordinary Go tests for the affected package before editing.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must update the
[decision register](docs/specification-decisions.md) and follow the
[specification governance contract](AGENTS.md#design), update
the affected stable decision entries, and complete the Specification Decisions
section of the pull request template. An unresolved interpretation or stale
source pin is release-blocking; peer behavior cannot silently select policy.

Selected mutation gates must finish with zero surviving viable mutants.

Do not fork shared verification workflows, add permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Use Go 1.27.2 for development and CI; both modules retain their Go 1.27.0
minimum. An ordinary checkout is sufficient for tests and a focused active fuzz
smoke, without sibling repositories or goal ledgers:

```bash
GOWORK=off go test ./...
GOWORK=off go test . -run '^$' -fuzz '^FuzzStructuredSources$' -fuzztime=2s
(cd adapters/awssecretsmanager && GOWORK=off go test ./...)
```

Select the native target and package affected by your change; ordinary Go tests
already exercise deterministic fuzz seeds. For shared checks, install the public
`golib` v2.1.1 release declared in `.golib.yaml` and verify its checksums. The
ordinary shared workflow and specification setup use that same public verifier:

```bash
make check
make ci
```

Both Make routes use bounded local checks. Neither module declares required
services or test tags. Optional mutation analysis can select one affected package:

```bash
golib mutation --module . --package ./decode
```

Use `make check-full` for explicit full assurance, including the online
specification check, configured full fuzz budgets, conformance and both modules'
API baselines. Hosted CI retains its separate blocking online specification job.
For source, filesystem, cancellation or AWS adapter changes, select the affected
behavioral, lifecycle and integration tests; the AWS adapter's client and retry
ownership remain caller-controlled. Run directly affected public consumers for
contract changes. Explicit full and release qualification retain their selected
gates and release rehearsal. The separate development security diagnostic is not
a release substitute. Report unavailable or failing checks without claiming
release readiness. CI/configuration-only changes do not require a library release.

## Selected Scheduled Assurance

The separate [selected-assurance caller](.github/workflows/assurance.yml) runs
only on main, every Tuesday at 04:11 UTC. It rotates one current declared root
fuzz target for 30 seconds, exploring:
structured source parsing, bounded interpolation and environment mapping,
destination decoding, filesystem inputs and discovery containment.
Only the root module's six declared targets participate; the AWS Secrets
Manager adapter's separate fuzz, API and integration contracts remain explicit.

Base-free ordinary checks execute deterministic seeds, not active exploration;
committed-range CI selects affected active smoke. Manual assurance requires an
exact anchored selector from the root fuzz operation in `.golib.yaml`, for
example `^FuzzDiscoveryBoundary$`.
Empty, unknown and malformed selections refuse.
The shared job bounds setup and execution to ten minutes, active workers and
CPU to two, owned memory to 2 GiB, native output to 1 MiB and private-cache
lifetime. Failure, cancellation, exhausted budgets and cleanup refusal remain
attributable failures. This lane does not run mutation or replace Required,
explicit `make check-full`, or relevant release qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
