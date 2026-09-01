# Changelog

All notable changes are documented here. The format follows Keep a Changelog
and releases use Semantic Versioning.

## [Unreleased]

### Changed

- Adopt the pinned `go-library-tools` v1.2.0 CLI and reusable workflow so CI
  enforces specification decisions, conformance bindings, source monitoring,
  and change control while retaining both module API baselines and approved
  mutation evidence.

### Documentation

- Replace archived monorepo and hardening terminology with package-owned
  documentation and reproducible verification guidance.

- Document the existing source-specific non-finite number behavior: JSON and
  YAML reject non-finite values while TOML preserves them as `float64`.

### Specification Decisions

- CONFIG-DEC-001 sha256:36c7c04c194ee53c5bcdba16b1a6b4d5771d1861a803de4b6cf1c84b04891b2e:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-002 sha256:bed8bd0f19425b5b3617e1921cc48c1f9c19df231028a63dff7db71187740d6c:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-003 sha256:f385bb784dbae76cdeb81164ecebeaf049de48d4592658b91ddcca2ba511bfb6:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-004 sha256:22415740993978504b16e29b9e79552a5a1b324cbae37cd5e0019e07d24762fc:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-005 sha256:6c70bad2d8e8808ae5a386975b7b92ebab98a5c7d4b34381c396f37e9e4f81e5:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-006 sha256:5fde9d85597a07907d9f816696ecf44ed23a402273ef1a2fb8c77283d70a6866:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-007 sha256:caa480104735facc30a670fe01ce79eeb175852a6564eebb96a54b80b896a189:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-008 sha256:ae3324209deb904ddbccad2a5e16805bb6195623d362d8f44611e3b8ea91668f:
  [Decision register](docs/specification-decisions.md).
- CONFIG-DEC-009 sha256:a7daf89cc315651001c22b15137caa0c90e254d0e60ae246d64f20da249bd794:
  [Decision register](docs/specification-decisions.md).

## [1.0.0] - 2026-08-25

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Replace obsolete standalone-repository links and workflow claims with
  monorepo-canonical targets and current release guidance.

- Link the package README to package-owned documentation.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-config` identity while preserving its documented API and behavior.
- Replace obsolete owned-module pseudo-version pins with the monorepo's local
  `v0.0.0` source-proxy coordinates; release tooling continues to emit exact
  `v1.0.0` dependency versions.
- Remove unused CLI-related indirect dependencies from canonical module
  metadata.
- Pin owned sibling modules to exact resolvable main pseudo-versions so
  standalone and clean external consumers use immutable dependency content.

- Unsigned typed defaults and environment or dotenv values within the signed
  integer range now use the same canonical numeric representation as JSON,
  YAML, and TOML sources, allowing equivalent higher-precedence values to
  override them without a false type conflict.
- Replaced obsolete package-local workflow references with the authoritative
  root CI matrix and its canonical per-module quality contract.
- Optional filesystem-backed sources now recognize missing files returned by
  `os.DirFS` without invoking extension-defined error callbacks.

### Added

- A `configservice` adapter that supplies typed, validated configuration to a
  selected service command before component construction, with explicit local
  dotenv and process-environment precedence.
- Typed immutable configuration plans and snapshots with safe provenance.
- Strict JSON, YAML, TOML, dotenv, environment, filesystem, defaults, and
  programmatic sources.
- Explicit bounded discovery, merge semantics, validation, optional values,
  redacted secrets, byte sizes, test helpers, fuzzing, race tests, and
  benchmarks.
- Exact coverage, compatibility, security, documentation, and release gates.
- Canonical source-tree enforcement, private mutable-state rejection, and
  secret-safe cause wrappers across every public parser and conversion error.
- Runnable source-composition examples plus filesystem and discovery fuzz
  targets for paths, policies, symlinks, permissions, and malformed data.
- Redacted custom-source and oversized numeric failures, plus deterministic
  cycle, depth, key-count, array-element, and cancellation guards for source
  trees and typed schemas.
- Complete precedence/merge permutation evidence and a structured-format
  conformance matrix with executable intentional-difference cases.
- `ErrSourceChanged`, context-aware filesystem/decode extension contracts, and
  generation checks for fail-closed reads under concurrent mutation.
- Diagnostic JSON/log canary coverage that seals arbitrary causes, plus
  Windows-compatible filtering of unrelated process-environment names.
- Exact case-sensitive discovery containment with filesystem-identity boundary
  checks, plus default rejection of Windows junction and mount-point reparse
  traversal.
- Required hardening traceability from every audit contract to executable
  evidence, including all 13,699 non-empty precedence subset permutations.
- A `configtest.DiffSecrets` assertion helper for value-aware diffs that never
  expose either secret operand.
- Explicit conformance assertions that decoding, defaults, environment
  loading, metadata, and snapshot cloning do not mutate private struct state.

[Unreleased]: https://github.com/faustbrian/go-config/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/faustbrian/go-config/releases/tag/v1.0.0
