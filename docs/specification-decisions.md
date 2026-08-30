# Specification decisions

This register records the observable JSON, YAML, and TOML interpretation
choices owned by `go-config`. General parser behavior is not treated as policy
until the package selects, rejects, or normalizes it at its public source boundary.

## CONFIG-DEC-001: Structured sources require one object-shaped document

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format; RFC 8259; section 2; https://www.rfc-editor.org/rfc/rfc8259.txt |
| Classification | omission; application-policy; requirement strength not specified |
| Issue | JSON permits any serialized value and YAML permits streams containing arbitrary root nodes, while configuration loading requires one mergeable document shape. |
| Credible interpretations | Accept every specification-valid root and define format-specific merge rules. Require one object, mapping, or root table and reject additional documents or root values. |
| Known peer behavior | General-purpose JSON and YAML decoders commonly accept scalar, sequence, and multi-document inputs; TOML always exposes a root table. |
| Selected behavior | Each structured source produces exactly one map[string]any root. JSON scalars, arrays, null, and trailing roots fail; YAML requires exactly one mapping document; TOML uses its root table. |
| Normative rationale | A single object-shaped document gives merge and decode one deterministic key-addressable input without inventing scalar or multi-document precedence. |
| Security consequences | Rejecting additional roots prevents hidden trailing configuration and avoids partial selection from multi-document streams. |
| Resource consequences | Only one bounded document is converted and retained. |
| Compatibility consequences | Previously rejected non-object roots remain errors across structured sources. |
| Wire-format consequences | Accepted source bytes are not rewritten; the restriction applies to the loaded configuration tree. |
| Executable tests | TestSourceRejectsEveryInvalidRootAndDelimiterState, TestSourceRejectsNonMappingRoot, TestBytesLoadsDottedKeysAndArrayTables |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | json.Bytes, json.FromFS, yaml.Bytes, yaml.FromFS, toml.Bytes, toml.FromFS |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | No upstream issue exists because the root restriction is application policy. |
| Reconsider when | The merge model gains an explicit scalar, sequence, or multi-document composition contract. |
Additional authoritative source: `{"id":"yaml-1.2.2-source","version":"YAML 1.2.2","url":"https://yaml.org/spec/1.2.2/","specifications":["YAML 1.2.2"]}`
Additional authoritative source: `{"id":"toml-1.0.0-source","version":"TOML 1.0.0","url":"https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md","specifications":["TOML 1.0.0"]}`

Structured contract:

- `omission`
- `application-policy`
- `RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format`
- `RFC 8259`
- `json-rfc8259-source`
- `https://www.rfc-editor.org/rfc/rfc8259.txt`
- `2`
- `not specified`
- `TestSourceRejectsEveryInvalidRootAndDelimiterState`
- `TestSourceRejectsNonMappingRoot`
- `TestBytesLoadsDottedKeysAndArrayTables`
- `FuzzStructuredSources`
- `json.Bytes`
- `json.FromFS`
- `yaml.Bytes`
- `yaml.FromFS`
- `toml.Bytes`
- `toml.FromFS`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-002: Duplicate member definitions fail closed

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format; RFC 8259; section 4; https://www.rfc-editor.org/rfc/rfc8259.txt |
| Classification | interoperability policy; recommended; requirement strength SHOULD |
| Issue | JSON only recommends unique object names and peer decoders disagree on first-wins, last-wins, all-values, or rejection behavior; YAML and TOML require unique mapping or key definitions. |
| Credible interpretations | Keep the first definition. Keep the last definition. Expose every duplicate. Reject the document before publishing a tree. |
| Known peer behavior | encoding/json keeps the last duplicate object value, while YAML and TOML parsers vary in diagnostics and rejection timing. |
| Selected behavior | JSON, YAML, and TOML reject duplicate definitions. JSON and YAML report a safe key path, YAML also reports location, and no partial tree is published. |
| Normative rationale | Failing closed follows JSON's interoperability recommendation and preserves YAML and TOML uniqueness instead of silently selecting authority by parser order. |
| Security consequences | An attacker cannot smuggle a shadow value that different readers interpret differently. |
| Resource consequences | Duplicate detection uses the same bounded key accounting as normal conversion. |
| Compatibility consequences | Duplicate configuration remains invalid rather than acquiring first-wins or last-wins semantics. |
| Wire-format consequences | No ambiguous document produces a normalized tree. |
| Executable tests | TestSourceRejectsDuplicateKeys, TestSourceRejectsAmbiguousOrExecutableFeatures, TestSourceRejectsDuplicateDefinitions |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | json.DuplicateKeyError, yaml.DuplicateKeyError, toml.Bytes |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | RFC 8259 documents the peer divergence; no package-specific upstream issue exists. |
| Reconsider when | A versioned compatibility mode explicitly exposes a different duplicate-member policy. |
Additional authoritative source: `{"id":"yaml-1.2.2-source","version":"YAML 1.2.2","url":"https://yaml.org/spec/1.2.2/","specifications":["YAML 1.2.2"]}`
Additional authoritative source: `{"id":"toml-1.0.0-source","version":"TOML 1.0.0","url":"https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md","specifications":["TOML 1.0.0"]}`

Structured contract:

- `interoperability policy`
- `recommended`
- `RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format`
- `RFC 8259`
- `json-rfc8259-source`
- `https://www.rfc-editor.org/rfc/rfc8259.txt`
- `4`
- `SHOULD`
- `TestSourceRejectsDuplicateKeys`
- `TestSourceRejectsAmbiguousOrExecutableFeatures`
- `TestSourceRejectsDuplicateDefinitions`
- `FuzzStructuredSources`
- `json.DuplicateKeyError`
- `yaml.DuplicateKeyError`
- `toml.Bytes`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-003: Shared numeric values use bounded Go scalar types

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format; RFC 8259; section 6; https://www.rfc-editor.org/rfc/rfc8259.txt |
| Classification | implementation-defined behavior; application-policy; requirement strength not specified |
| Issue | The formats define numeric syntax but do not select the Go types, precision, or overflow behavior of one cross-format configuration tree. |
| Credible interpretations | Convert every number to float64. Preserve parser-specific numeric types. Use arbitrary-precision values. Normalize integers to int64 then uint64 and finite fractional values to float64. |
| Known peer behavior | encoding/json defaults to float64 unless UseNumber is enabled; YAML and TOML decoders expose source-specific numeric types. |
| Selected behavior | Signed integers normalize to int64, larger non-negative integers through math.MaxUint64 normalize to uint64, finite fractional and exponent values normalize to float64, and values outside the supported domain fail. |
| Normative rationale | The bounded scalar set preserves common integer precision, supports typed decoding, and keeps equivalent finite documents equal across formats. |
| Security consequences | Overflow and non-finite conversion errors fail without including the sensitive numeric token. |
| Resource consequences | Numeric conversion has fixed memory and does not allocate arbitrary-precision values. |
| Compatibility consequences | The normalized Go scalar types and overflow boundaries remain part of the source contract. |
| Wire-format consequences | Numeric source syntax is consumed into a canonical tree and is not serialized back to the original format. |
| Executable tests | TestSourceConvertsNullNegativeAndUnsignedNumberBoundaries, TestSourceRejectsNumberOverflowCategories, TestScalarAndIntegerBoundaries, TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | Document.Tree, json.Bytes, yaml.Bytes, toml.Bytes |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | No upstream issue exists because Go scalar selection is package-owned. |
| Reconsider when | The public tree model adopts an explicit arbitrary-precision numeric value type. |
Additional authoritative source: `{"id":"yaml-1.2.2-source","version":"YAML 1.2.2","url":"https://yaml.org/spec/1.2.2/","specifications":["YAML 1.2.2"]}`
Additional authoritative source: `{"id":"toml-1.0.0-source","version":"TOML 1.0.0","url":"https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md","specifications":["TOML 1.0.0"]}`

Structured contract:

- `implementation-defined behavior`
- `application-policy`
- `RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format`
- `RFC 8259`
- `json-rfc8259-source`
- `https://www.rfc-editor.org/rfc/rfc8259.txt`
- `6`
- `not specified`
- `TestSourceConvertsNullNegativeAndUnsignedNumberBoundaries`
- `TestSourceRejectsNumberOverflowCategories`
- `TestScalarAndIntegerBoundaries`
- `TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree`
- `FuzzStructuredSources`
- `Document.Tree`
- `json.Bytes`
- `yaml.Bytes`
- `toml.Bytes`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-004: Non-finite numbers remain source-language specific

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | TOML 1.0.0; TOML 1.0.0; section Float; https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md |
| Classification | interoperability policy; application-policy; requirement strength not specified |
| Issue | JSON forbids Infinity and NaN, YAML core resolution can produce non-finite floats, and TOML 1.0.0 explicitly defines inf and nan while the common tree uses float64. |
| Credible interpretations | Reject non-finite values in every format. Accept every parser's non-finite values. Preserve TOML's normative values while keeping JSON and the defensive YAML profile finite-only. |
| Known peer behavior | TOML decoders normally expose inf and nan as floating-point values; JSON parsers reject them and YAML parsers vary by schema and safety profile. |
| Selected behavior | JSON rejects non-finite syntax, YAML rejects resolved infinity and NaN scalars, and TOML preserves inf, -inf, and nan as float64 values. |
| Normative rationale | The policy preserves valid TOML 1.0.0 data without broadening JSON grammar or the deliberately finite YAML configuration profile. |
| Security consequences | Callers decoding TOML floating-point values must handle non-finite values explicitly; other formats cannot introduce them. |
| Resource consequences | Non-finite TOML values use ordinary float64 storage. |
| Compatibility consequences | TOML callers retain accepted non-finite values while JSON and YAML remain finite-only. |
| Wire-format consequences | The distinction is source-format specific and visible only in the normalized Go tree. |
| Executable tests | TestSourcePreservesTOMLNonFiniteFloats, TestScalarAndIntegerBoundaries, TestSourceRejectsNumberOverflowCategories |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | toml.Bytes, yaml.Bytes, json.Bytes, Document.Tree |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | No upstream issue exists; the package intentionally chooses a narrower YAML profile and preserves TOML 1.0.0 values. |
| Reconsider when | The canonical tree adopts a format-independent finite-number restriction. |
Additional authoritative source: `{"id":"json-rfc8259-source","version":"RFC 8259","url":"https://www.rfc-editor.org/rfc/rfc8259.txt","specifications":["RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format"]}`
Additional authoritative source: `{"id":"yaml-1.2.2-source","version":"YAML 1.2.2","url":"https://yaml.org/spec/1.2.2/","specifications":["YAML 1.2.2"]}`

Structured contract:

- `interoperability policy`
- `application-policy`
- `TOML 1.0.0`
- `TOML 1.0.0`
- `toml-1.0.0-source`
- `https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md`
- `Float`
- `not specified`
- `TestSourcePreservesTOMLNonFiniteFloats`
- `TestScalarAndIntegerBoundaries`
- `TestSourceRejectsNumberOverflowCategories`
- `FuzzStructuredSources`
- `toml.Bytes`
- `yaml.Bytes`
- `json.Bytes`
- `Document.Tree`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-005: YAML graph and application tag features are disabled

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | YAML 1.2.2; YAML 1.2.2; section 3.2.1.2, 3.2.2.2, 3.3.4, and 7.1; https://yaml.org/spec/1.2.2/ |
| Classification | optional behavior; defensive; requirement strength not specified |
| Issue | YAML supports aliases, anchors, application tags, complex keys, and graph structures that do not map safely or unambiguously into the package's JSON-shaped configuration tree. |
| Credible interpretations | Expose the complete YAML representation graph. Expand aliases and merge keys. Allow registered application tags. Accept only core scalar, sequence, and string-keyed mapping nodes without aliases. |
| Known peer behavior | General-purpose YAML loaders commonly expand aliases and may resolve application tags or merge keys; safety-focused configuration loaders often disable them. |
| Selected behavior | YAML accepts core scalar, sequence, and mapping nodes with string keys, but rejects aliases, merge keys, custom tags, non-string keys, tagged collections, and multiple documents. |
| Normative rationale | The bounded tree model has no identity, executable tag, or complex-key semantics, so accepting those features would create hidden expansion and format-only behavior. |
| Security consequences | Disabling aliases and application tags prevents expansion attacks, object construction, and merge-key shadowing. |
| Resource consequences | Conversion work is bounded by explicit depth and key limits without alias expansion. |
| Compatibility consequences | Documents using advanced YAML graph or tag features remain rejected and must be rewritten as explicit core values. |
| Wire-format consequences | Accepted YAML is normalized to an acyclic map and slice tree. |
| Executable tests | TestSourceRejectsAmbiguousOrExecutableFeatures, TestConvertRejectsMalformedAndUnsupportedNodes, TestSourceRejectsNonMappingRoot |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | yaml.Bytes, yaml.FromFS, yaml.ParseError |
| Documentation | docs/conformance.md, docs/sources.md, docs/security.md, docs/specification-decisions.md |
| Upstream status | The YAML specification permits application-specific availability decisions; no package-specific upstream issue exists. |
| Reconsider when | A separate opt-in YAML graph API defines identity, expansion, tag registration, and resource ownership. |

Structured contract:

- `optional behavior`
- `defensive`
- `YAML 1.2.2`
- `YAML 1.2.2`
- `yaml-1.2.2-source`
- `https://yaml.org/spec/1.2.2/`
- `3.2.1.2, 3.2.2.2, 3.3.4, and 7.1`
- `not specified`
- `TestSourceRejectsAmbiguousOrExecutableFeatures`
- `TestConvertRejectsMalformedAndUnsupportedNodes`
- `TestSourceRejectsNonMappingRoot`
- `FuzzStructuredSources`
- `yaml.Bytes`
- `yaml.FromFS`
- `yaml.ParseError`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/security.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-006: YAML timestamps normalize to inert strings

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | YAML 1.2.2; YAML 1.2.2; section 3.3.2 and 10.3; https://yaml.org/spec/1.2.2/ |
| Classification | implementation-defined behavior; application-policy; requirement strength not specified |
| Issue | YAML tag resolution and parser compatibility behavior can surface timestamp-shaped scalars as native time values even though the common configuration tree has no timezone or calendar value type. |
| Credible interpretations | Expose parser-native time.Time values. Reject timestamp-resolved scalars. Normalize the resolved lexical value to a string. |
| Known peer behavior | YAML libraries differ by schema and compatibility mode on whether timestamp-shaped plain scalars become strings or native date-time values. |
| Selected behavior | A YAML scalar resolved as a timestamp is returned as its lexical string value without timezone conversion or clock interpretation. |
| Normative rationale | Strings preserve inert configuration data and align timestamp-shaped values with JSON while avoiding implicit calendar semantics. |
| Security consequences | Loading configuration cannot trigger timezone lookup or time-based side effects. |
| Resource consequences | Timestamp normalization allocates only the bounded scalar string. |
| Compatibility consequences | YAML timestamp-shaped values remain strings in Document.Tree. |
| Wire-format consequences | The lexical timestamp is preserved as a string rather than converted to a canonical time instant. |
| Executable tests | TestScalarAndIntegerBoundaries, TestCrossFormatDifferencesAreIntentionalAndNormalized |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | yaml.Bytes, yaml.FromFS, Document.Tree |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | go.yaml.in/yaml v4 documents preserved YAML 1.1 compatibility behavior; no package-specific upstream issue exists. |
| Reconsider when | The public tree introduces an explicit date-time value with documented timezone semantics. |

Structured contract:

- `implementation-defined behavior`
- `application-policy`
- `YAML 1.2.2`
- `YAML 1.2.2`
- `yaml-1.2.2-source`
- `https://yaml.org/spec/1.2.2/`
- `3.3.2 and 10.3`
- `not specified`
- `TestScalarAndIntegerBoundaries`
- `TestCrossFormatDifferencesAreIntentionalAndNormalized`
- `FuzzStructuredSources`
- `yaml.Bytes`
- `yaml.FromFS`
- `Document.Tree`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-007: TOML date and time values normalize to strings

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | TOML 1.0.0; TOML 1.0.0; section Offset Date-Time, Local Date-Time, Local Date, and Local Time; https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md |
| Classification | implementation-defined behavior; application-policy; requirement strength not specified |
| Issue | TOML defines four date and time types, but the common configuration tree must choose whether to preserve parser-native time.Time values, distinguish TOML categories, or expose strings. |
| Credible interpretations | Expose time.Time and parser locations. Create four package-specific temporal types. Normalize each TOML category to a stable lexical string. |
| Known peer behavior | BurntSushi TOML represents all four categories with time.Time plus synthetic locations for local categories. |
| Selected behavior | Offset date-time values use RFC3339Nano strings; local date-time, local date, and local time use fixed lexical layouts without adding a timezone. |
| Normative rationale | String normalization preserves category meaning, avoids synthetic location leakage, and aligns date-time-shaped values with JSON and YAML configuration. |
| Security consequences | Configuration loading performs no timezone lookup or clock-dependent conversion. |
| Resource consequences | Normalization uses bounded fixed-layout string formatting. |
| Compatibility consequences | All TOML date and time categories remain strings in Document.Tree. |
| Wire-format consequences | Date and time values are normalized to stable strings and are not serialized back to TOML. |
| Executable tests | TestSourceNormalizesEveryTOMLDateTimeCategory, TestCrossFormatDifferencesAreIntentionalAndNormalized |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | toml.Bytes, toml.FromFS, Document.Tree |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | No upstream issue exists because the mapping from TOML temporal values to the package tree is application policy. |
| Reconsider when | The public tree introduces explicit local and offset temporal value types. |

Structured contract:

- `implementation-defined behavior`
- `application-policy`
- `TOML 1.0.0`
- `TOML 1.0.0`
- `toml-1.0.0-source`
- `https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md`
- `Offset Date-Time, Local Date-Time, Local Date, and Local Time`
- `not specified`
- `TestSourceNormalizesEveryTOMLDateTimeCategory`
- `TestCrossFormatDifferencesAreIntentionalAndNormalized`
- `FuzzStructuredSources`
- `toml.Bytes`
- `toml.FromFS`
- `Document.Tree`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-008: Equivalent structured documents share one canonical tree

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format; RFC 8259; section 2 through 6; https://www.rfc-editor.org/rfc/rfc8259.txt |
| Classification | interoperability policy; application-policy; requirement strength not specified |
| Issue | JSON, YAML, and TOML have overlapping but non-identical data models, so parser-native outputs can make precedence and typed decoding depend on the source format. |
| Credible interpretations | Expose each parser's native value graph. Normalize only at typed decode time. Normalize the shared object, array, string, boolean, and finite-number subset before merge and document language-specific gaps. |
| Known peer behavior | Format-specific libraries expose distinct integer, timestamp, table-array, null, and graph representations. |
| Selected behavior | Equivalent objects, arrays of objects, strings, booleans, and finite numbers normalize identically before merge. JSON and YAML null become nil; TOML has no null. Timestamp and TOML date-time values normalize to strings. |
| Normative rationale | Pre-merge normalization makes source substitution predictable while preserving explicit differences that the source languages cannot share. |
| Security consequences | Format substitution cannot silently activate aliases, tags, or parser-native object types. |
| Resource consequences | Normalization copies only bounded trees and does not retain parser graphs. |
| Compatibility consequences | Equivalent documents continue to produce deeply equal Document.Tree values across formats. |
| Wire-format consequences | The package defines a canonical in-memory tree, not a canonical serialized wire format. |
| Executable tests | TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree, TestJSONYAMLAndTOMLArrayOfObjectsProduceSameTree, TestCrossFormatDifferencesAreIntentionalAndNormalized |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | Document.Tree, LoadTree, json.Bytes, yaml.Bytes, toml.Bytes |
| Documentation | docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | No upstream standard defines the package's cross-format tree; this is an application interoperability contract. |
| Reconsider when | A new structured source cannot map its shared values without weakening an existing normalized type or invariant. |
Additional authoritative source: `{"id":"yaml-1.2.2-source","version":"YAML 1.2.2","url":"https://yaml.org/spec/1.2.2/","specifications":["YAML 1.2.2"]}`
Additional authoritative source: `{"id":"toml-1.0.0-source","version":"TOML 1.0.0","url":"https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md","specifications":["TOML 1.0.0"]}`

Structured contract:

- `interoperability policy`
- `application-policy`
- `RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format`
- `RFC 8259`
- `json-rfc8259-source`
- `https://www.rfc-editor.org/rfc/rfc8259.txt`
- `2 through 6`
- `not specified`
- `TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree`
- `TestJSONYAMLAndTOMLArrayOfObjectsProduceSameTree`
- `TestCrossFormatDifferencesAreIntentionalAndNormalized`
- `FuzzStructuredSources`
- `Document.Tree`
- `LoadTree`
- `json.Bytes`
- `yaml.Bytes`
- `toml.Bytes`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## CONFIG-DEC-009: Structured parsing is bounded and diagnostics are redacted

| Field | Decision |
| --- | --- |
| Status and owner | resolved; go-config maintainers |
| Source | RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format; RFC 8259; section 9 and 12; https://www.rfc-editor.org/rfc/rfc8259.txt |
| Classification | omission; defensive; requirement strength not specified |
| Issue | The format specifications permit implementations to set limits but do not define this library's byte, depth, key, cancellation, diagnostic, or secret-redaction boundaries. |
| Credible interpretations | Delegate all limits and errors to parser defaults. Apply only a byte limit. Enforce common byte, depth, and key limits with cancellation and redact parser details. |
| Known peer behavior | General-purpose parsers expose detailed syntax errors and format-specific resource behavior; their defaults are not a shared hostile-input contract. |
| Selected behavior | Each structured source enforces caller-configurable byte, depth, and key limits, checks cancellation during owned work, publishes no partial tree, and exposes stable errors without source tokens. |
| Normative rationale | Configuration may contain secrets and can be attacker-controlled, so bounded work and redacted failure surfaces are package-owned requirements. |
| Security consequences | Hostile inputs cannot request unbounded owned traversal or leak source contents through public diagnostics. |
| Resource consequences | Default limits are 1 MiB, depth 64, and 100000 keys, with caller-selected positive overrides. |
| Compatibility consequences | Inputs beyond configured limits and detailed dependency error messages remain unavailable. |
| Wire-format consequences | Rejected input produces no normalized document. |
| Executable tests | TestSourceEnforcesExactDepthKeyAndArrayIndexBoundaries, TestConvertEnforcesExactDepthKeyAndSequenceBoundaries, TestNormalizeEnforcesExactDepthKeyAndCollectionBoundaries, TestErrorsHaveStableSecretSafeFormatting, TestParseErrorHasStableSecretSafeFormatting, TestSecretAndErrorsNeverLeakAcrossDiagnosticSurfaces |
| Fixture evidence | None applicable. |
| Fuzz evidence | FuzzStructuredSources |
| Interoperability evidence | Not assessed; no official cross-format fixture or provider defines this application tree. |
| Differential evidence | Not assessed; parser dependencies are implementation inputs, not independent maintained peers for the package policy. |
| Public APIs | json.Limits, yaml.Limits, toml.Limits, yaml.ParseError, toml.ParseError |
| Documentation | docs/security.md, docs/conformance.md, docs/sources.md, docs/specification-decisions.md |
| Upstream status | RFC 8259 allows implementation limits; YAML and TOML do not define the package's common resource and diagnostic policy. |
| Reconsider when | A parser backend changes owned cancellation or traversal boundaries, or the public limit model changes. |
Additional authoritative source: `{"id":"yaml-1.2.2-source","version":"YAML 1.2.2","url":"https://yaml.org/spec/1.2.2/","specifications":["YAML 1.2.2"]}`
Additional authoritative source: `{"id":"toml-1.0.0-source","version":"TOML 1.0.0","url":"https://raw.githubusercontent.com/toml-lang/toml/1.0.0/toml.md","specifications":["TOML 1.0.0"]}`

Structured contract:

- `omission`
- `defensive`
- `RFC 8259 The JavaScript Object Notation (JSON) Data Interchange Format`
- `RFC 8259`
- `json-rfc8259-source`
- `https://www.rfc-editor.org/rfc/rfc8259.txt`
- `9 and 12`
- `not specified`
- `TestSourceEnforcesExactDepthKeyAndArrayIndexBoundaries`
- `TestConvertEnforcesExactDepthKeyAndSequenceBoundaries`
- `TestNormalizeEnforcesExactDepthKeyAndCollectionBoundaries`
- `TestErrorsHaveStableSecretSafeFormatting`
- `TestParseErrorHasStableSecretSafeFormatting`
- `TestSecretAndErrorsNeverLeakAcrossDiagnosticSurfaces`
- `FuzzStructuredSources`
- `json.Limits`
- `yaml.Limits`
- `toml.Limits`
- `yaml.ParseError`
- `toml.ParseError`
- `docs/security.md`
- `docs/conformance.md`
- `docs/sources.md`
- `docs/specification-decisions.md`

## Unresolved decisions

None. New source formats, parser dialect changes, or changes to any registered
normalization, rejection, resource, or diagnostic boundary require a new decision
or a superseding entry before implementation.
