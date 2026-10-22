# Changelog

All notable changes follow Keep a Changelog and semantic versioning.

## 3.0.0 - 2026-10-22

### Changed

- Move to `github.com/faustbrian/go-config/adapters/awssecretsmanager/v3` and
  build on root `github.com/faustbrian/go-config/v3`, so sources satisfy the
  root-v3 `config.Source`. Secret retrieval, payload bounds and error
  classification are unchanged.

## 2.1.0 - 2026-10-22

Published from the notes below, which earlier releases left unversioned.

### Documentation

- Clarify SDK/client ownership, post-materialization payload admission,
  transient-copy erasure limits and direct provider cause exposure.

### Changed

- Prepare the adapter's independent v2 identity and Go 1.27 minimum. Use
  `github.com/faustbrian/go-config/adapters/awssecretsmanager/v3` with root-v2
  plans and loaders; returned sources no longer have the root-v1 type identity.

- Adopt schema-v2 cohesion metadata and the repository's checksum-verified
  v1.4.0 W14 gate without changing the adapter API or runtime behavior.

- Reconcile the root module's v1.0.0 checksum with its immutable public module
  identity so clean consumers verify the adapter dependency successfully.

### Documentation

- Link the adapter's stable-v1 support, exact Go compatibility, canonical tag
  form, and live repository support and private-reporting routes from the
  shared package documentation.

- Add canonical installation and import guidance, the exact Go 1.26.6 support
  floor, and explicit construction, lifecycle, concurrency, resource,
  security, compatibility, and ecosystem navigation contracts.

- Clarify the cancellation-aware client invocation bound, document portable-Go
  support, the 2,048-byte secret-identifier bound, and actionable
  troubleshooting; remove obsolete unreleased status text and correct the
  v1.0.0 release date.

- Publish the adapter's selection, ownership, lifecycle, support, and delivery
  boundaries and link its entry point to the immutable v1.4.0 ecosystem index.

- Add a module documentation index for direct navigation.
- Use human-oriented section names and package-owned documentation links.

## 2.0.1 - 2026-10-11

### Changed

- Update AWS Secrets Manager to 1.50.4 with the matching SDK and Smithy
  graph so received responses are closed before returning interceptor errors.
  Preserve sensitive JSON results, private error causes, and caller-owned
  retries with one tracing loop per operation.

## 1.0.0 - 2026-08-26

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
