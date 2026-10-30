# Changelog

All notable changes are documented here. The format follows Keep a Changelog
and releases use Semantic Versioning.

## Unreleased

### Added

- Add `SourceDigest`, `Document.Digest` and `Snapshot.SourceDigest(name)`.
  The JSON, YAML, TOML and dotenv sources, including their filesystem forms,
  report the SHA-256 of the exact bytes each load accepted, so callers can bind
  a loaded value to the bytes it came from without reading the file twice.

### Changed

- Release the AWS Secrets Manager adapter as
  `adapters/awssecretsmanager/v3.0.0`, built on root v3. See its
  [changelog](adapters/awssecretsmanager/CHANGELOG.md).

## 3.0.0 - 2026-10-22

### Changed

- Move the root module to `github.com/faustbrian/go-config/v3` and bind
  `adapters/service` to `github.com/faustbrian/go-service/v3`, whose command
  types build on go-cli v2. `Loader[T]` now takes a Service v3 `Invocation`.

### Removed

- Remove the `configservice` compatibility facade. Import
  `github.com/faustbrian/go-config/v3/adapters/service` instead; its options,
  loader, `ErrInvalidOptions` and `OptionsError` are the same, but its
  `OptionsError` is the only structured option error.

## 2.0.2 - 2026-10-22

### Changed

- Publish the current source and documentation with a fresh immutable tag.
- Build with Go 1.27.3 to address GO-2026-6809 in `crypto/x509`.
  Consumers should use Go 1.27.3 or newer.
- Publish independently versioned modules from the same source snapshot:
  - `adapters/awssecretsmanager/v2.1.0`.

## Earlier unversioned notes (shipped by 2.0.1)

### Fixed

- Restore independent typed snapshot copying for mutable references inside
  arrays and validate Optional payloads before cloning or validator invocation.
- Reject cyclic programmatic containers and typed hook results before recursive
  owned traversal. Preserve finite recursive instances and overlapping acyclic
  slices without adding new constructor depth or count limits.
- Consume deletion markers at keys in newly copied nested objects; preserve the
  programmatic source's separate structural admission policy.
- Recheck cancellation after the final validator or lone self-validator so
  canceled candidates are not published.

### Documentation

- Clarify per-source admission, typed snapshot/wrapper copying limitations,
  trusted metadata and cooperative callbacks/I/O; reconcile current
  CONFIG-DEC-001/002/009 parsing and diagnostic prose.

### Changed

- Prepare the root v2 module identity and Go 1.27 minimum. Migrate imports to
  `github.com/faustbrian/go-config/v2`; existing v1 tags retain their original
  module identity and Go support contract.

### Security

- Confine `filesystem.FromDiscovered` loads to the approved directory identity
  and reject discovery-to-open symlink or directory replacement races instead
  of reopening a swapped target outside the trusted root.

### Changed

- Update Service and Correlation to v1.1.0 while retaining the service
  adapter and compatibility facade loader APIs, configuration precedence,
  cancellation, and safe error behavior.

- Adopt Service v1.1.3 with bounded maintenance-file reads and
  Correlation v1.1.1 with redacted entropy failures while preserving the
  typed service adapter and compatibility facade loader contracts.

## [2.0.1] - 2026-10-13

Source range: `v2.0.0` (`81b390daf5`) through `2b79cffba8`.

### Fixed

- Restore independent snapshot copying through mutable array elements and
  Optional payload admission before cloning or callbacks. Reject actual runtime
  cycles while preserving finite recursive instances and shared acyclic
  containers. Consume nested object tombstones and recheck cancellation after
  the final validator so canceled candidates are not published.
  [2b79cffba8](https://github.com/faustbrian/go-config/commit/2b79cffba8b705c05d6b403be12d068c37162565).

### Changed

- Adopt Service v1.1.3 and Correlation v1.1.1 while retaining service-loader,
  precedence, cancellation and error contracts.
  [6da5222d48](https://github.com/faustbrian/go-config/commit/6da5222d481507dcb0c9a5cf962701e0b358274f).

### Documentation

- Clarify source/typed ownership, admission, specification boundaries and
  current public adoption/security scope without changing selected standards.
  [cbdd055819](https://github.com/faustbrian/go-config/commit/cbdd055819c9e4532d2a0c90faf05c06fcfffa59), [889a65ca91](https://github.com/faustbrian/go-config/commit/889a65ca919ae33206d36ed80de8ceb02af304ed), [104627c13d](https://github.com/faustbrian/go-config/commit/104627c13dd24c640096346f0ca994cdc2292a7f).

### Maintenance

- Keep repository policy and authenticated tooling aligned; retain bounded
  assurance, source-documentation checks and credential-free specification
  checkout. These changes do not add root runtime behavior.
  [8063c591c8](https://github.com/faustbrian/go-config/commit/8063c591c87c40c11e7a8dbcf542c350d1f713f0), [f0284a48c7](https://github.com/faustbrian/go-config/commit/f0284a48c7dc6dcbd8e1c20648e0515a519e1a11), [25f2fbaeec](https://github.com/faustbrian/go-config/commit/25f2fbaeecf04abcb1708ae0f55ad0afc14844b2), [933a164244](https://github.com/faustbrian/go-config/commit/933a164244858695a927e98d1ee9366a55ab4d4c), [deed76a214](https://github.com/faustbrian/go-config/commit/deed76a21471e4070867fd5fb0644a06660467c8), [42a7ee4f0f](https://github.com/faustbrian/go-config/commit/42a7ee4f0ff2645d4d4865509e3b68d47d1384d1), [6a7e0bf364](https://github.com/faustbrian/go-config/commit/6a7e0bf364d373bb1e8431b78b50a3872509b8d9), [0f35d06d63](https://github.com/faustbrian/go-config/commit/0f35d06d6328655e636ebf314891994ec7c6cbd4), [ae0ee73192](https://github.com/faustbrian/go-config/commit/ae0ee731922e2c7ed0688280a54dd51f826108e0), [b4e93530e3](https://github.com/faustbrian/go-config/commit/b4e93530e3b0a119422a511791fab1d1b29c508d).
- Account for the independently released AWS adapter's v2 identity and SDK
  response-cleanup maintenance in this repository range. Those changes belong
  to its existing releases, not this root patch; no new adapter release or
  dependency repin is required. See its [changelog](adapters/awssecretsmanager/CHANGELOG.md).
  [f03e83d52d](https://github.com/faustbrian/go-config/commit/f03e83d52d4e7250238974f564f2ef558f92ff2f), [675e166198](https://github.com/faustbrian/go-config/commit/675e1661981669e61146e158ea55ba036d019223), [1399afc70a](https://github.com/faustbrian/go-config/commit/1399afc70ae36337d23b7e4b43ddf8c9ce5d35ea).

## [1.1.0] - 2026-09-09

### Added

- Add the target-oriented `adapters/service` import path for typed service
  command loading while retaining `configservice` as an API-compatible facade.

### Changed

- Adopt the checksum-verified `go-library-tools` v1.6.2 CLI, schema-v2
  cohesion metadata for both modules, repository-local cohesion gate, and
  proportional shared workflow without changing public APIs or runtime behavior.

- Reconcile Golib v1.0.0 dependency checksums with their immutable public
  module identities so clean consumers verify released dependencies
  successfully.

- Adopt the pinned `go-library-tools` v1.2.0 CLI and reusable workflow so CI
  enforces specification decisions, conformance bindings, source monitoring,
  and change control while retaining both module API baselines and approved
  mutation evidence.
  The required specification lane remains independent of module-resolution
  failures in the broader repository gate.

### Documentation

- Document stable root and AWS adapter entry points, canonical v1 installation,
  exact Go support, caller-owned lifecycle boundaries, unambiguous tag forms,
  complete navigation, and live support and security routes.

- Publish each module's family, capabilities, selection, ownership, lifecycle,
  support, and delivery boundaries and link both entry points to the immutable
  v1.4.0 ecosystem index and Foundations guidance.

- Align public minimum-Go and CI-platform statements with the module's Go
  1.26.6 directive and authoritative Ubuntu workflow.

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

[2.0.1]: https://github.com/faustbrian/go-config/compare/v2.0.0...v2.0.1
[1.1.0]: https://github.com/faustbrian/go-config/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/faustbrian/go-config/releases/tag/v1.0.0
