# Changelog

All notable changes follow Keep a Changelog and semantic versioning.

## Unreleased

### Changed

- Adopt schema-v2 cohesion metadata and the repository's checksum-verified
  v1.4.0 W14 gate without changing the adapter API or runtime behavior.

- Reconcile the root module's v1.0.0 checksum with its immutable public module
  identity so clean consumers verify the adapter dependency successfully.

### Documentation

- Add canonical installation and import guidance, the exact Go 1.26.6 support
  floor, and explicit construction, lifecycle, concurrency, resource,
  security, compatibility, and ecosystem navigation contracts.

- Clarify the cancellation-aware client invocation bound, document portable-Go
  support and actionable troubleshooting, and remove obsolete unreleased
  status text after the v1.0.0 adapter release.

- Publish the adapter's selection, ownership, lifecycle, support, and delivery
  boundaries and link its entry point to the immutable v1.4.0 ecosystem index.

- Add a module documentation index for direct navigation.
- Use human-oriented section names and package-owned documentation links.

## 1.0.0 - 2026-08-25

### Documentation

- Link the package README to package-owned documentation.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-config/adapters/awssecretsmanager` identity while preserving its documented API and behavior.
- Replace the obsolete owned-module pseudo-version pin with the monorepo's
  local `v0.0.0` source-proxy coordinate; release tooling continues to emit
  the exact `v1.0.0` dependency version.

### Added

- A bounded, secret-safe AWS Secrets Manager JSON configuration source with
  explicit version selection, optional-source semantics, and caller-owned AWS
  default credential-chain composition.
