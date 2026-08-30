# Structured-format specification conformance

The [specification decision register](../docs/specification-decisions.md)
defines the application profile layered over RFC 8259, YAML 1.2.2, and TOML
1.0.0. The source formats remain authoritative for their syntax; the register
owns only `go-config` normalization, rejection, resource, and diagnostic policy.

`sources.tsv` pins the exact reviewed normative publications. `monitoring.json`
separately pins RFC errata and the official YAML and TOML specification tag
feeds. Parser dependencies are implementation inputs and are not treated as
normative sources or independent peers.

## Decision bindings

| Decision | Executable evidence |
| --- | --- |
| CONFIG-DEC-001: Structured sources require one object-shaped document | `TestSourceRejectsEveryInvalidRootAndDelimiterState`, `TestSourceRejectsNonMappingRoot`, `TestBytesLoadsDottedKeysAndArrayTables` |
| CONFIG-DEC-002: Duplicate member definitions fail closed | `TestSourceRejectsDuplicateKeys`, `TestSourceRejectsAmbiguousOrExecutableFeatures`, `TestSourceRejectsDuplicateDefinitions` |
| CONFIG-DEC-003: Shared numeric values use bounded Go scalar types | `TestSourceConvertsNullNegativeAndUnsignedNumberBoundaries`, `TestSourceRejectsNumberOverflowCategories`, `TestScalarAndIntegerBoundaries`, `TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree` |
| CONFIG-DEC-004: Non-finite numbers remain source-language specific | `TestSourcePreservesTOMLNonFiniteFloats`, `TestScalarAndIntegerBoundaries`, `TestSourceRejectsNumberOverflowCategories` |
| CONFIG-DEC-005: YAML graph and application tag features are disabled | `TestSourceRejectsAmbiguousOrExecutableFeatures`, `TestConvertRejectsMalformedAndUnsupportedNodes`, `TestSourceRejectsNonMappingRoot` |
| CONFIG-DEC-006: YAML timestamps normalize to inert strings | `TestScalarAndIntegerBoundaries`, `TestCrossFormatDifferencesAreIntentionalAndNormalized` |
| CONFIG-DEC-007: TOML date and time values normalize to strings | `TestSourceNormalizesEveryTOMLDateTimeCategory`, `TestCrossFormatDifferencesAreIntentionalAndNormalized` |
| CONFIG-DEC-008: Equivalent structured documents share one canonical tree | `TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree`, `TestJSONYAMLAndTOMLArrayOfObjectsProduceSameTree`, `TestCrossFormatDifferencesAreIntentionalAndNormalized` |
| CONFIG-DEC-009: Structured parsing is bounded and diagnostics are redacted | `TestSourceEnforcesExactDepthKeyAndArrayIndexBoundaries`, `TestConvertEnforcesExactDepthKeyAndSequenceBoundaries`, `TestNormalizeEnforcesExactDepthKeyAndCollectionBoundaries`, `TestErrorsHaveStableSecretSafeFormatting`, `TestParseErrorHasStableSecretSafeFormatting`, `TestSecretAndErrorsNeverLeakAcrossDiagnosticSurfaces` |

## Differential boundary

No maintained peer currently exposes the same bounded `map[string]any`
application contract across all three source languages. Dependency agreement
would be circular evidence because the YAML and TOML parsers are inputs to this
implementation. Cross-format equality is therefore verified as package
conformance, while maintained-peer differential evidence remains `not assessed`.
