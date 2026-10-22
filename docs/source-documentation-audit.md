# Source documentation audit

[Documentation index](README.md)

This is a historical audit of the v2 source at the base commit below. Rows
naming `configservice`, Service v1 or the `/v2` module describe that base, not
the current v3 tree.

## Owner, contract and bounded checks

Base ae0ee731922e2c7ed0688280a54dd51f826108e0; detached main-source worktree.
Tier A documentation-only. The repository owns deterministic low-to-high
configuration composition, typed snapshots, explicit source adapters and loading
boundaries; applications own source trust, names/locations, deployment, refresh
scheduling and borrowed SDK/filesystem/callback resources. Direct consumers are
Plan/Load callers, standalone source and decoding users, config service adapters
and the separately versioned AWS Secrets Manager module.

Preserve executable tokens/literals/AST/directives, dependencies, API baselines,
fixtures, source authorities and historical decision prefixes. Audit all 85 Go
files across two modules and 22 catalog packages (including three example
commands, test support and two internal packages). Native structural API mode is
apidiff for both modules, not full Go-doc bytes; preserve both baseline files.
No generator or generated Go header is present. Inspect actual scripts/CI before
choosing commands; no generator is executed for comment-count enforcement.

Named risks: precedence and tombstone semantics; source tree admission versus
whole-process limits; defensive cloning and typed-state rejection; sensitive
values versus metadata/cause exposure; callbacks/context/SDK I/O cooperativity;
filesystem containment and generation checks; service publication/failure windows;
parser interpretation and pinned JSON/YAML/TOML decisions. Smallest checks:
changed formatting; all-source token/literal/AST/directive identity; public/member
and canonical-header structure; parsed comment HTML and all package renders;
local source/docs/offline specification and affected links; frozen-input and
changed/staged pinned-secret checks; one ordinary full final-diff inspection.
No unchanged tests/examples, active fuzz, race/leak, benchmark, external AWS,
consumer, online source or release campaign is selected. Go runs wait for an
explicit coordinator slot and use three immediately disposable owned caches.

## Missing versus investigated documentation

Initially missing and repaired: public field/interface method contracts, canonical package
boundaries, source metadata trust, callback/resource ownership, snapshot zero-value
and copying limitations, precise loader and cancellation/failure guarantees.
Existing meaningful comments remain. Priority constants needed name-led contracts,
not just a collective label. No TODO/FIXME/HACK/generated/deprecation markers were
found; platform constraints cover Windows discovery links and Darwin/Linux file
change metadata, and must remain unchanged.

Investigated source limitations, not runtime repair: typed cloning handles pointers,
maps/slices and exported struct fields but not mutable references inside arrays;
Optional clone hooks bypass payload validation. Source canonicalization counters
reset per source, with no total-plan byte/source/merged-key budget. Key sorting
and backend YAML/TOML materialization can allocate before owned structural checks;
constructors copy before load byte admission. Programmatic construction requires
trusted acyclic containers and does not accept merge.Delete; nested tombstones in
newly copied subtrees are not recursively processed by merge. Source names,
locations/paths and Origin overrides are declared metadata, not sanitized or
authenticated. Standalone decode/merge and validator callbacks are not input or
I/O sandboxes, and successful final validators are not followed by another
context check. Source/SDK/factory callbacks may block/panic at uncovered seams;
Reader uses ordinary Close directly while fs-backed sourceio uses independent
one-second cooperative context close. Generation checks detect observed trusted
metadata changes, not every intervening write or content authenticity. Discovery
results do not hold descriptors and OwnerOnly checks mode bits, not UID/ACLs.

Direct JSON syntax errors retain standard-library causes that may reveal offending
characters; plan-level SourceError hides arbitrary source text/types. AWS direct
errors retain raw SDK causes intentionally, while plan loading redacts them.
Payload admission follows SDK materialization/local copy; clearing that copy does
not erase strings/parser/SDK/decoded allocations. Sensitive marks provenance,
not protected storage. Runtime defects/stronger requirements remain maintainer
follow-ups, not silently repaired or newly passing behavior.

Pinned source authorities: RFC 8259, YAML 1.2.2, TOML 1.0.0; JSON uses stdlib,
YAML dependency is v4.0.0-rc.6, TOML v1.6.0. Root consumes go-service v1.1.3;
AWS uses service/secretsmanager v1.50.4 and published root/v2 v2.0.0, not a local
replace. Current narrow record changes are CONFIG-DEC-001/002/009 only; actual
typed history digests are appended without changing stable IDs/source pins or
old prefixes. Nine records remain current. Both apidiff baseline bytes stay frozen.
All tests/helpers/fuzz/bench/examples/platform files are source evidence, not
newly executed tests. No TestMain or service/test tag requirement is declared.
No tracked scripts directory or source/API producer exists; native Make/CI gates
are not universal documentation qualification.

## Reproducible complete module/package/file inventory

Use `git ls-files '*.go' '*go.mod' '*go.sum'`, modules.json/packages.json,
package/declaration scans, `.golib.yaml`, native Make/CI and marker scans.
Root module is github.com/faustbrian/go-config/v2; nested module is
github.com/faustbrian/go-config/adapters/awssecretsmanager/v2. Both minimum Go
versions are 1.27.0; repository verification pins 1.27.2. The complete paths below
include public/private implementations, tests/helpers, seven declared native fuzz
boundaries, benchmarks, executable examples and all platform variants.

| Go file | Package | Audit role and existing evidence |
| --- | --- | --- |
| `adapters/awssecretsmanager/benchmark_test.go` | `awssecretsmanager` | Existing test/helper evidence: BenchmarkSourceLoad |
| `adapters/awssecretsmanager/example_test.go` | `awssecretsmanager_test` | Existing test/helper evidence: ExampleNew |
| `adapters/awssecretsmanager/fuzz_test.go` | `awssecretsmanager` | Existing test/helper evidence: FuzzSourceLoad |
| `adapters/awssecretsmanager/response_cleanup_test.go` | `awssecretsmanager_test` | Existing test/helper evidence: TestSourceClosesSDKResponsesAndKeepsOneRetryLoop |
| `adapters/awssecretsmanager/source.go` | `awssecretsmanager` | Source contract and important implementation invariants |
| `adapters/awssecretsmanager/source_test.go` | `awssecretsmanager` | Existing test/helper evidence: TestSourceLoadsSensitiveJSONFromCurrentSecretVersion, TestSourceLoadsBinaryJSONAtExplicitImmutableVersion, TestSourceAcceptsExactResourceLimits, TestSourcePreservesEachExplicitVersionSelector, TestNewRejectsUnsafeComposition, TestSourceRejectsInvalidProviderResponses, TestSourceMapsMissingSecretAndRedactsProviderFailures, TestSourceHonorsCancellationAndPropagatesSafeJSONFailures |
| `adapters/service/example_test.go` | `configservice_test` | Existing test/helper evidence: ExampleNew |
| `adapters/service/service.go` | `configservice` | Source contract and important implementation invariants |
| `adapters/service/service_test.go` | `configservice_test` | Existing test/helper evidence: TestLoaderAppliesLocalDotenvEnvironmentAndOverridePrecedence, TestLoaderUsesEnvironmentOverLocalDotenv, TestNewRejectsDotenvOutsideExplicitLocalMode, TestNewRejectsInvalidSourceConstruction, TestOptionsErrorUnwrapIncludesOnlyPresentCause, TestNewRejectsInvalidSourcePlan, TestLoaderPreservesCancellationAndValidationCauses, TestLoaderSnapshotsOptionSlicesAndReadsProcessEnvironmentPerLoad |
| `benchmarks_test.go` | `config_test` | Existing test/helper evidence: BenchmarkSourcePlanLoad, BenchmarkDecode, BenchmarkMerge, BenchmarkValidation, BenchmarkLargeConfig |
| `bytesize.go` | `config` | Source contract and important implementation invariants |
| `bytesize_test.go` | `config_test` | Existing test/helper evidence: TestByteSizeParsesSupportedUnitsAndPlainBytes, TestByteSizeRejectsNegativeMalformedAndOverflowValues |
| `concurrency_test.go` | `config_test` | Existing test/helper evidence: TestPlanLoadsDeterministicallyInParallel, TestConcurrentPlansHooksValidatorsSnapshotsAndFailures |
| `config.go` | `config` | Source contract and important implementation invariants |
| `configservice/configservice.go` | `configservice` | Source contract and important implementation invariants |
| `configservice/configservice_test.go` | `configservice_test` | Existing test/helper evidence: TestCompatibilityFacadePreservesLoaderAndErrorIdentity |
| `configservice/example_test.go` | `configservice_test` | Existing test/helper evidence: ExampleNew |
| `configtest/configtest.go` | `configtest` | Source contract and important implementation invariants |
| `configtest/configtest_test.go` | `configtest_test` | Existing test/helper evidence: TestNewSourceIsDeterministicAndImmutable, TestFailingSourcePreservesCause, TestEnvironmentReturnsSortedSnapshot, TestFilesystemReturnsIndependentReadableFiles, TestMustLoadAndAssertOrigin, TestDiffSecretsReportsOnlyRedactedValues, TestSourceHonorsCancellationAndClonesOrigins, TestHelpersReportPlanLoadAndOriginFailures |
| `decode/decode.go` | `decode` | Source contract and important implementation invariants |
| `decode/decode_test.go` | `decode_test` | Existing test/helper evidence: TestIntoRecoversAndRedactsTextUnmarshalerPanic, TestIntoAggregatesIndependentFieldErrorsDeterministically, TestIntoDecodesTypedStruct, TestIntoRejectsUnknownFieldWithoutMutatingDestination, TestIntoReportsNestedTypeErrorSafely, TestIntoRejectsInvalidDestination, TestValueDecodesPointersNullsAndInterfacesAtomically, TestValueDecodesNumericKindsAndRejectsBoundaries, TestIntoReportsCollectionFailuresByElementPath, TestIntoRejectsAmbiguousIgnoredAndUnsupportedFields, TestIntoReportsScalarHookFailuresSafely, TestTypedErrorsExposeSafeMetadataAndCauses, TestHookCauseIsRedactedAtEveryExposedErrorLayer, TestValueCoversSuccessfulHooksNilScalarsAndPointerFailures, TestValueAcceptsEveryIntegerSourceKind, TestContextDecodeHooksObserveCancellationAndRecoverPanics |
| `decode/wrapped_errors_test.go` | `decode` | Existing test/helper evidence: TestFlattenErrorsPreservesWrappedStructuredFailures |
| `default_plan.go` | `config` | Source contract and important implementation invariants |
| `default_plan_test.go` | `config_test` | Existing test/helper evidence: TestNewDefaultPlanAssignsDocumentedPrecedence, TestNewDefaultPlanDoesNotMutateSourceMetadata, TestNewDefaultPlanRejectsDuplicateNamesAcrossCategories, TestNewDefaultPlanRejectsTypedNilSourceWithoutPanicking |
| `defaults/defaults.go` | `defaults` | Source contract and important implementation invariants |
| `defaults/defaults_test.go` | `defaults_test` | Existing test/helper evidence: TestForCanonicalizesUnsignedDefaultsLikeStructuredSources, TestForBuildsTypedDefaultSource, TestForRejectsInvalidDefaultAndDestination, TestForRejectsRecursiveSchemaWithoutRecursingIndefinitely, TestForConvertsOptionalScalarDefault, TestForConvertsPointerOptionalScalarDefault, TestForRejectsTrailingCollectionDefaultInput, TestDefaultSourceDoesNotExposeNestedCollectionState, TestDefaultErrorsAreTypedAndDoNotExposeInput, TestForConvertsEverySupportedDefaultShape, TestForRejectsInvalidNestedStructCollectionAndUnsupportedDefaults |
| `defaults/internal_hardening_test.go` | `defaults` | Existing test/helper evidence: TestCollectRejectsSchemaBeyondDepthLimitInternally, TestCollectContinuesAcrossNonDefaultFieldsInternally |
| `diagnostic_redaction_test.go` | `config_test` | Existing test/helper evidence: TestSecretAndErrorsNeverLeakAcrossDiagnosticSurfaces |
| `discover/discover.go` | `discover` | Source contract and important implementation invariants |
| `discover/discover_hardening_test.go` | `discover` | Existing test/helper evidence: TestValidateRejectsEveryPolicyAndPathCategory, TestValidatePropagatesInjectedPlatformFailures, TestDiscoveryPlatformFailuresPreserveIdentityWithoutLeakingText, TestValidateUsesDefaultUserConfigurationDirectory, TestAbsoluteJoinsRootAndRejectsEscapes, TestSearchHandlesDuplicateMissingAndBoundedResults, TestSearchRejectsExplicitAndDirectoryPathsOutsideRoot, TestSearchRejectsSymlinkLoopsWithoutExposingPaths, TestSearchFollowsHostFilesystemPathCasing, TestSearchStopsAtCanonicalCaseVariantDirectory, TestSearchPropagatesUpwardCanonicalizationFailure, TestInspectUsesPathComponentsForSymlinkDetection, TestRootlessSymlinkWalkInspectsAncestors, TestSymlinkWalkStopsBeforeInspectingConfiguredRoot, TestSearchPropagatesInjectedCandidateAndExplicitStatFailures, TestSearchPropagatesVisitErrorsFromEveryTraversalKind, TestSearchStopsInsideExplicitDirectory, TestSearchEnforcesUpwardDepthAndCanStopAtStart, TestSearchPropagatesCancellationDuringCandidateConsideration, TestInspectHandlesMissingDirectoryBrokenLinkAndNonRegularFile, TestWithinRejectsParentAndAcceptsSameOrChild, TestAbsoluteAndInspectPropagateInjectedPlatformFailures |
| `discover/discover_test.go` | `discover_test` | Existing test/helper evidence: TestSearchDoesNotTraverseParentsByDefault, TestSearchUpwardStopsAtExplicitDirectory, TestSearchHonorsDirectoryAndPlaceOrder, TestSearchAcceptsExplicitFilesAndDirectories, TestSearchUserConfigDirectoryIsExplicit, TestSearchEnforcesSymlinkAndRootPolicies, TestSearchChecksSecretFilePermissions, TestSearchRejectsUnsafeOrUnboundedOptions, TestSearchHonorsCancellationAndCandidateLimit, TestSearchAllowsExactResultLimitAndContinuesAfterExplicitDirectory |
| `discover/fuzz_test.go` | `discover` | Existing test/helper evidence: FuzzDiscoveryBoundary |
| `discover/link_bits.go` | `discover` | Source contract and important implementation invariants |
| `discover/link_bits_test.go` | `discover` | Existing test/helper evidence: TestLinkFlagHelpersRequireTheRequestedBit |
| `discover/link_other.go` | `discover` | Source contract and important implementation invariants |
| `discover/link_windows.go` | `discover` | Source contract and important implementation invariants |
| `discover/link_windows_test.go` | `discover` | Existing test/helper evidence: TestPathContainsSymlinkRejectsWindowsReparseComponent, TestIsLinkLikeClassifiesSymlinksAndWindowsAttributes |
| `dotenv/dotenv.go` | `dotenv` | Source contract and important implementation invariants |
| `dotenv/dotenv_hardening_test.go` | `dotenv` | Existing test/helper evidence: TestUnquotedEscapeAdvancesPastEscapedCharacter, TestErrorsHaveStableSecretSafeFormatting, TestConstructorsValidateInputsAndPreserveImmutableOptions, TestFSSourcePropagatesOpenReadCloseLimitAndCancellation, TestParseRejectsEverySyntaxAndLimitCategory, TestParseReportsExactSyntaxLocationsAndAllowsLimitBoundaries, TestValueParsersReportExactLinesColumnsAndCommentBoundaries, TestParsingCoversEscapesCommentsNamesAndCancellation, TestInterpolationCoversResolutionFallbackAndBounds, TestInterpolationAllowsExactBoundsAndClassifiesEmptyExpressions, TestLoadPropagatesMapperAndInterpolationErrors |
| `dotenv/dotenv_test.go` | `dotenv_test` | Existing test/helper evidence: TestBytesParsesDocumentedDotenvGrammar, TestBytesRejectsDuplicateKeys, TestInterpolationIsExplicitBoundedAndCycleDetected, TestSourceEnforcesLimits, TestFSSourceOptionalOnlySuppressesMissingFile, TestFSSourcePreservesPermissionErrors, TestEnvironmentWinsOverDotenvByDefault, TestSourceHonorsCanceledContext |
| `environment/environment.go` | `environment` | Source contract and important implementation invariants |
| `environment/environment_hardening_test.go` | `environment_test` | Existing test/helper evidence: TestProcessForReadsFreshProcessSnapshotWithoutMutation, TestEnvironForIgnoresUnrelatedWindowsEnvironmentNames, TestEnvironForRejectsInvalidOptionsAndSchemas, TestEnvironForMapsImplicitPointerIgnoredAndUnicodeFields, TestEnvironForHonorsCancellationBetweenEntries, TestEnvironmentErrorsAreTypedStableAndSecretSafe, TestEnvironForConvertsEveryCollectionElementShape, TestEnvironForConvertsUnsignedScalar, TestEnvironForAllowsExactLimitsAndClassifiesMalformedWantedVariable, TestEnvironForRejectsEveryInvalidCollectionShape, TestEnvironForRejectsScalarAndCollectionConversionFailures, TestEnvironForIgnoresUnrelatedPlatformVariableNames |
| `environment/environment_test.go` | `environment_test` | Existing test/helper evidence: TestEnvironForMapsAndConvertsTypedFields, TestEnvironForRejectsMissingRequiredField, TestMappingErrorFormatsAndMarshalsOnlyItsSafeMessage, TestLaterSourceMaySupplyRequiredEnvironmentField, TestEnvironForRejectsDuplicateAndNormalizedCollisions, TestEnvironForRejectsSchemaCollisions, TestEnvironForRejectsRecursiveSchemaWithoutRecursingIndefinitely, TestEnvironForDoesNotMutateInput, TestEnvironForRejectsMalformedVariableAndValue, TestEnvironForHonorsCancellation, TestEnvironForEnforcesExplicitResourceLimits, TestEnvironForConvertsOptionalScalarValue, TestEnvironForConvertsPointerOptionalScalarValue, TestEnvironForRejectsTrailingCollectionInput, TestEnvironForConvertsNumericAndBooleanCollections |
| `environment/internal_hardening_test.go` | `environment` | Existing test/helper evidence: TestCollectFieldsRejectsSchemaBeyondDepthLimitInternally, TestCollectFieldsContinuesAcrossSkippedAndNestedFieldsInternally, TestNormalizeAndValidNameBoundariesInternally |
| `examples/discovery/main.go` | `main` | Source contract and important implementation invariants |
| `examples/quickstart/main.go` | `main` | Source contract and important implementation invariants |
| `examples/testing/main.go` | `main` | Source contract and important implementation invariants |
| `examples_test.go` | `config_test` | Existing test/helper evidence: Example_structuredSources, Example_dotenvSources, Example_environmentSources, Example_defaultsAndProgrammaticSources, Example_filesystemSources, Example_discoveryAndValidation, Example_secretsProvenanceAndConfigtest |
| `filesystem/filesystem.go` | `filesystem` | Source contract and important implementation invariants |
| `filesystem/filesystem_hardening_test.go` | `filesystem` | Existing test/helper evidence: TestReaderValidatesEveryOptionBoundary, TestReaderDispatchesYAMLAndTOMLAndPreservesMetadata, TestReaderRejectsNilFactoryResultAndMalformedDocument, TestFromFSRejectsInvalidPathFormatAndLimits, TestFromDiscoveredValidatesPathsAndFallsBackToLexicalPath, TestPathConstructorsPropagateAbsolutePathFailure, TestNewDiscoveredFSRejectsUnavailableOrNonDirectoryRoot, TestDiscoveredFSRejectsOpenBoundaryFailures, TestLocationSourcePreservesExistingOriginsAndMarksNull, TestReaderSourceRejectsInternallyInvalidFormat, TestReaderSourceHonorsAlreadyCanceledContext |
| `filesystem/filesystem_test.go` | `filesystem_test` | Existing test/helper evidence: TestFromFSDispatchesSupportedExtensions, TestFromFSRejectsUnsupportedFormatUnlessExplicit, TestFromPathReloadsAtomicallyReplacedFile, TestFromDiscoveredPreservesPathProvenance, TestFromDiscoveredRejectsTargetSwapOutsideApprovedDirectory, TestFromDiscoveredReloadsAtomicallyReplacedRegularFile, TestFromDiscoveredRejectsApprovedDirectoryReplacement, TestFromDiscoveredRejectsRemovedApprovedDirectory, TestOptionalFileSuppressesOnlyAbsence, TestFromFSValidatesOptions, TestReaderOpensFreshContextAwareInputForEveryLoad, TestReaderEnforcesReadAndLifecycleFailures, TestReaderHonorsCancellationAndByteLimit, TestReaderRequiresFactoryAndExplicitFormat |
| `filesystem/fuzz_test.go` | `filesystem` | Existing test/helper evidence: FuzzFilesystemBoundary |
| `format_equivalence_test.go` | `config_test` | Existing test/helper evidence: TestUnsignedTypedDefaultCanBeOverriddenByEquivalentJSONNumber, TestUnsignedTypedDefaultCanBeOverriddenByEquivalentDotenvNumber, TestUnsignedTypedDefaultPreservesValueAboveSignedRange, TestJSONYAMLAndTOMLEquivalentDocumentsProduceSameTree, TestJSONYAMLAndTOMLArrayOfObjectsProduceSameTree, TestCrossFormatDifferencesAreIntentionalAndNormalized, TestFilesystemBackedFormatsRejectMutationDuringRead |
| `fuzz_test.go` | `config_test` | Existing test/helper evidence: FuzzStructuredSources, FuzzDotenvInterpolation, FuzzEnvironmentMapping, FuzzDecodeTagsAndDestinationTypes |
| `internal/safeerror/safeerror.go` | `safeerror` | Source contract and important implementation invariants |
| `internal/safeerror/safeerror_test.go` | `safeerror_test` | Existing test/helper evidence: TestRedactPreservesIdentityWithoutExposingCause |
| `internal/sourceio/change_darwin.go` | `sourceio` | Source contract and important implementation invariants |
| `internal/sourceio/change_linux.go` | `sourceio` | Source contract and important implementation invariants |
| `internal/sourceio/change_other.go` | `sourceio` | Source contract and important implementation invariants |
| `internal/sourceio/sourceio.go` | `sourceio` | Source contract and important implementation invariants |
| `internal/sourceio/sourceio_test.go` | `sourceio` | Existing test/helper evidence: TestBytesIsImmutableRepeatableAndBounded, TestContextReaderChecksContextAndSystemChangeToken, TestFromFSValidatesAndReopensFiles, TestFromFSPreservesNonAbsenceOpenError, TestFromFSClassifiesMissingOSDirFSFileAsNotFound, TestFromFSClosesFileReturnedWithOpenError, TestFromFSPreservesHostileMissingErrorWithoutCallbacks, TestFromFSPreservesReaderErrorWithoutCallingIs, TestFromFSContainsHostileFilesystemOperations, TestFromFSUsesContextAwareOperationsWhenAvailable, TestFromFSClosesWithIndependentContextAfterCancellation, TestFromFSRejectsFileMutationDuringRead, TestStableFSReadRejectsSameMetadataGenerationChange, TestStableFSReadRejectsInvalidFileAndStatFailures, TestInputReadHandlesReadAndCloseFailures, TestReadValidatesBoundsErrorsAndCancellation, TestFileChangedDistinguishesGenerationState, TestClassifyNotExistHandlesEveryWrapperBoundary |
| `json/json.go` | `json` | Source contract and important implementation invariants |
| `json/json_hardening_test.go` | `json_test` | Existing test/helper evidence: TestFromFSValidatesFilesystemPathAndOptions, TestBytesAreImmutableAndPreserveSourceMetadata, TestFSSourcePropagatesReadCloseLimitAndCancellationErrors, TestSourceRejectsEveryInvalidRootAndDelimiterState, TestSourceClassifiesTrailingRootTokensExactly, TestSourceEnforcesExactDepthKeyAndArrayIndexBoundaries, TestSourceConvertsNullNegativeAndUnsignedNumberBoundaries, TestSourceRejectsNumberOverflowCategories, TestDuplicateKeyErrorFormattingIsStable, TestSourceHonorsCancellationDuringParsing |
| `json/json_test.go` | `json_test` | Existing test/helper evidence: TestBytesLoadsStrictTree, TestBytesRejectsInvalidOptions, TestSourceRejectsDuplicateKeys, TestOversizedNumberDoesNotExposeSensitiveToken, TestSourceEnforcesLimits, TestFSSourceOptionalOnlySuppressesMissingFile, TestFSSourcePreservesPermissionErrors, TestSourceHonorsCanceledContext |
| `merge/merge.go` | `merge` | Source contract and important implementation invariants |
| `merge/merge_test.go` | `merge_test` | Existing test/helper evidence: TestTrees, TestTreesRejectsTypeChanges, TestTreesReportsTheLexicallyFirstConflict, TestTreesContinuesAfterRecursivelyMergingAnObject, TestTreesDoesNotMutateInputs, TestTreesReplacesInheritedNullWithTypedValue, TestTreesReportsNestedObjectAndSliceKinds, TestTreesCoversEveryRepresentativeKindPair |
| `metadata_test.go` | `config_test` | Existing test/helper evidence: TestLoadPropagatesSecretAndDeprecatedFieldMetadataToOrigins, TestSensitiveSourceMarksEveryNestedOriginWithoutRetainingValues |
| `optional.go` | `config` | Source contract and important implementation invariants |
| `optional_test.go` | `config_test` | Existing test/helper evidence: TestOptionalPreservesAbsentNullEmptyAndZero, TestOptionalCollectionsRemainImmutableAcrossSnapshotReads, TestOptionalPreservesWinningDefaultedState, TestOptionalDefaultedStateIsVisibleToValidators, TestOptionalRejectsInvalidTypedValueAtomically |
| `plan_test.go` | `config_test` | Existing test/helper evidence: TestNewPlanOrdersSourcesByPrecedence, TestNewPlanPreservesCallerOrderAtEqualPriority, TestNewPlanRejectsInvalidSourceMetadata, TestNewPlanRejectsTypedNilSourceWithoutPanicking, TestLoadTreeIsAtomicAndTracksWinningSource, TestLoadTreeReturnsNoSnapshotWhenLaterSourceFails, TestSnapshotDoesNotExposeMutableState, TestTypedSnapshotClonesExportedFieldsWhenStructHasPrivateState, TestLoadDecodesTypedImmutableSnapshot, TestLoadReturnsNoTypedSnapshotOnDecodeFailure, TestLoadDecodeErrorIncludesWinningSourceAndLocation, TestLoadRunsSelfValidationWithoutReturningFailedSnapshot, TestLoadWithValidatorsAggregatesAfterCompleteDecode |
| `precedence_audit_test.go` | `config_test` | Existing test/helper evidence: TestNewPlanEveryPrioritySubsetAndOrderHasTheDocumentedWinner, TestNewPlanEveryEqualPriorityPermutationPreservesCallerOrder, TestDefaultCompositionPreservesPresenceAndHighestPrecedenceWinner |
| `programmatic/programmatic.go` | `programmatic` | Source contract and important implementation invariants |
| `programmatic/programmatic_test.go` | `programmatic_test` | Existing test/helper evidence: TestDefaultsAndOverridesUseDocumentedPriorities, TestMapSourceDoesNotMutateOrExposeInput, TestMapRejectsMutablePointerValues, TestMapSourceRejectsInvalidNameAndHonorsCancellation, TestMapSourceNormalizesTypedCollectionsWithoutAliasing, TestMapSourceRejectsNonStringMapKeys, TestDefaultsMarkEveryNestedPathAsDefaulted, TestMapSourceNormalizesArraysNamedKeysAndTypedNilCollections, TestMapSourceReportsNestedInvalidMapPath |
| `root_hardening_test.go` | `config_test` | Existing test/helper evidence: TestLoadTreeOnlySuppressesAbsentOptionalSources, TestLoadTreeRedactsArbitrarySourceErrors, TestLoadTreeIsCanceledBeforeLoadingOrMerging, TestLoadTreeTracksNullDeletionAndOriginOverrides, TestLoadAnnotatesEveryDecodeFailureWithNearestOrigin, TestLoadRequiredFieldWithoutOriginRemainsSafelyUnattributed, TestLoadUsesParentOriginForMissingNestedField, TestSnapshotValueClonesAllPublicMutableShapes, TestLoadTreeRejectsMutableNonCanonicalSourceValues, TestLoadTreeRejectsCyclicAndBoundedSourceTrees, TestLoadTreeBoundsKeysAndHonorsCancellationDuringCanonicalization, TestLoadRejectsPrivateMutableStateWithoutSnapshotCloner |
| `root_internal_hardening_test.go` | `config` | Existing test/helper evidence: TestLoadWithValidatorsPropagatesTreeLoadFailureInternally, TestPresenceUsesDefaultAndCommaSuffixedFieldNamesInternally, TestOptionalReportsItsTypedTextTargetInternally, TestCloneReflectReconstructsPointerFromValueCloneInternally, TestNearestOriginRejectsIncompleteCollectionPathsInternally, TestCanonicalTreePropagatesNestedArrayValueErrorsInternally, TestCanonicalTreeChecksContextAtObjectBoundariesInternally, TestCanonicalTreeEnforcesExactBudgetsInternally, TestMetadataAndPresenceIgnoreExplicitlySkippedFieldsInternally, TestSnapshotCloneGuardsCoverNestedMutableShapesInternally |
| `secret.go` | `config` | Source contract and important implementation invariants |
| `secret_test.go` | `config_test` | Existing test/helper evidence: TestSecretRedactsFormattingAndMarshaling, TestSecretDecodesFromText, TestSecretDirectFormattingContractsAreRedacted, TestSecretComparisonsRemainRedactedWhenFormatted |
| `toml/toml.go` | `toml` | Source contract and important implementation invariants |
| `toml/toml_hardening_test.go` | `toml` | Existing test/helper evidence: TestParseErrorHasStableSecretSafeFormatting, TestConstructorsValidateInputsAndPreserveImmutableMetadata, TestFSSourcePropagatesReadCloseAndLimitErrors, TestSourceNormalizesEveryTOMLDateTimeCategory, TestNormalizeCollectionAndUnsupportedValueBoundaries, TestNormalizeEnforcesExactDepthKeyAndCollectionBoundaries, TestJoinHandlesRootAndNestedPaths, TestSourceHonorsCancellationDuringNormalization |
| `toml/toml_test.go` | `toml_test` | Existing test/helper evidence: TestBytesLoadsDottedKeysAndArrayTables, TestSourcePreservesTOMLNonFiniteFloats, TestSourceRejectsDuplicateDefinitions, TestParseErrorDoesNotExposeInput, TestSourceEnforcesLimits, TestFSSourceOptionalOnlySuppressesMissingFile, TestFSSourcePreservesPermissionAndCancellation |
| `validation/validation.go` | `validation` | Source contract and important implementation invariants |
| `validation/validation_test.go` | `validation_test` | Existing test/helper evidence: TestRunAggregatesFailuresInRegistrationOrder, TestRunErrorTextDoesNotExposeValidatorCause, TestRunHonorsCanceledContext, TestRunRecoversAndRedactsValidatorPanic, TestRunInvokesValueAndPointerSelfValidatorsBeforeRegisteredValidators, TestRunHandlesNilValidatorAndCancellationBetweenValidators, TestRunReturnsNilForValidCandidateWithoutValidators, TestRunRecoversAndRedactsSelfValidatorPanic, TestFailureFormattingAndAggregateUnwrapAreSafe, TestRunRedactsRawCustomCauseAtEveryExposedErrorLayer |
| `yaml/yaml.go` | `yaml` | Source contract and important implementation invariants |
| `yaml/yaml_hardening_test.go` | `yaml` | Existing test/helper evidence: TestErrorsHaveStableSecretSafeFormatting, TestConstructorsValidateInputsAndPreserveImmutableMetadata, TestSourceHandlesEmptyInputAndPropagatesFilesystemFailures, TestSourceRejectsNonMappingRoot, TestConvertRejectsMalformedAndUnsupportedNodes, TestConvertEnforcesExactDepthKeyAndSequenceBoundaries, TestScalarAndIntegerBoundaries, TestSourceHonorsCancellationDuringConversion |
| `yaml/yaml_test.go` | `yaml_test` | Existing test/helper evidence: TestBytesLoadsStrictTree, TestSourceRejectsAmbiguousOrExecutableFeatures, TestDuplicateErrorIncludesSafePathAndLocation, TestParseErrorDoesNotExposeInput, TestOversizedIntegerDoesNotExposeSensitiveToken, TestSourceEnforcesLimits, TestFSSourceOptionalOnlySuppressesMissingFile, TestFSSourcePreservesPermissionAndCancellation |

## Verification disposition

Executed locally under the explicit serialized slot with Go 1.27.2:
36 owned Go paths formatted; every one of the 85 tracked files has identical
position/comment-free AST, noncomment tokens/literals/implicit semicolons and
existing directive/marker text. No semicolon waiver. Both modules' go.mod/sum,
catalog/toolchain/configuration and structural apidiff baselines are unchanged.
All 22 catalog packages rendered with native go doc -all, including internal,
example-command, test-support and adapter packages. After documenting the public
anonymous-interface Origin method on AssertOrigin, that affected package was
rendered again; representative root/AWS/filesystem renders were inspected and
18 unaffected package renders reused. The final Map nil-root clarification was
then formatted/rendered at 1485 bytes, with all 85 AST/token/HTML checks and zero
source findings repeated; 21 unchanged package renders and nine typed/offline
specification proofs were reused. Parsed HTML covers 512 comment groups;
441 name-led public AST declarations/fields/methods plus the public parameter
method are documented, with 22 canonical package/command headers.

The native source checker reports zero findings for 85 Go files/two modules.
Local docs validation passes. Offline specification validation passes for the
one registered specification module/nine current decisions; the AWS module has
no specification register. Native typed Go canonicalization reproduces all nine
baseline digests and the three current appended CONFIG-DEC-001/002/009 digests;
old history prefixes, IDs, authorities and unaffected typed fields remain exact.
Both APIs use structural apidiff; no prose snapshot or source generator is run.

All executable example/test bodies, dependencies and fixtures are unchanged.
Unchanged tests/examples, active fuzz, race/leak, benchmarks, AWS/live services,
online specification monitoring, public consumers and release qualification
were not run. Existing test names and frozen evidence are source authorities,
not newly passing runtime proof. Failed preliminary source structure identified
one visible anonymous-interface method; its useful provenance contract was added
and the native check is now green, without suppressing findings.

Final lightweight validation covers 110 affected local link targets, all 85 audit
rows and 150 untouched tracked inputs, including both structural API baselines.
The three narrow current-record field deltas/history prefixes are checked.
Pinned v8.30.1 default-rule full owned-content, changed-plus-added and final
staged-diff secret checks have zero findings, without ignores or allowlist changes.
One ordinary complete final-diff truth/usefulness inspection covers all source,
maintained prose, current records/history and the audit. No independent Tier C/D
or runtime qualification is claimed. Native sessions, all exact three-cache directories, helpers and renders were removed
immediately on success/failure. Parent binaries and existing verification
archives remain untouched. The detached worktree stays for coordinator delivery;
this worker does not publish or claim hosted CI/release qualification.

Maintainer follow-ups, not runtime repairs: recursive array isolation and Optional
payload validation; stronger whole-plan/string/constructor/parser admission;
programmatic cycle/tombstone handling; final-callback context rechecks; uniform
factory/close panic/liveness cleanup; direct JSON syntax redaction when required;
source metadata/authenticity and SDK trust/policy. These are concrete source
limitations or stronger optional contracts, not guessed rationale or release proof.

## Objective Enforcement

The checksum-pinned public Tools v2.4.0 verifier now supplies
`make source-docs-check` as a prerequisite of the existing local and CI Make
routes. Ordinary shared CI selects strict source-documentation checking and
requires its result. Native prerequisites, operation budgets, security and
release controls are retained. Hosted exact-head evidence is recorded by the
coordination ledger; configuration alone does not establish green delivery.

Continue with [API reference](api.md).
