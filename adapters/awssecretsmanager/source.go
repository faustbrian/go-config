// Package awssecretsmanager loads explicit JSON configuration from AWS Secrets
// Manager through a caller-owned Client. New fixes options; each Load requests
// a fresh provider result and parses it as a sensitive config source.
// Payload size is checked after SDK materialization and an adapter copy, not as
// an HTTP memory cap. The transient copy is cleared but strings, parser copies,
// SDK responses and returned documents are not securely erased. Direct provider
// errors redact formatting but expose raw causes through Unwrap; config plan
// loading applies its separate redaction boundary. No credential discovery,
// retry loop, client close, caching, refresh or background goroutine is owned.
package awssecretsmanager

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	secretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	config "github.com/faustbrian/go-config/v3"
	jsonsource "github.com/faustbrian/go-config/v3/json"
)

const (
	// MaximumSecretBytes is the AWS Secrets Manager payload limit.
	MaximumSecretBytes = 65_536
	// MaximumSecretIDBytes bounds an AWS secret name or ARN before network IO.
	MaximumSecretIDBytes = 2_048
)

var (
	// ErrClientRequired identifies missing AWS Secrets Manager composition.
	ErrClientRequired = errors.New("AWS Secrets Manager client is required")
	// ErrInvalidOptions identifies an unsafe or incomplete source definition.
	ErrInvalidOptions = errors.New("AWS Secrets Manager config options are invalid")
	// ErrOperation identifies a failed provider read without exposing its details.
	ErrOperation = errors.New("AWS Secrets Manager config read failed")
	// ErrInvalidResponse identifies an absent, ambiguous, or oversized payload.
	ErrInvalidResponse = errors.New("AWS Secrets Manager config response is invalid")
)

// Client is the AWS Secrets Manager read surface, not an IAM enforcement boundary.
// It is borrowed and must honor context, return well-formed results and support
// the application's repeated/concurrent loads. The adapter does not catch panics.
type Client interface {
	// GetSecretValue performs the provider read; Client owns SDK retries and response cleanup.
	GetSecretValue(
		context.Context,
		*secretsmanager.GetSecretValueInput,
		...func(*secretsmanager.Options),
	) (*secretsmanager.GetSecretValueOutput, error)
}

// Options defines one explicit JSON configuration source. The default version
// stage is AWSCURRENT and the default maximum payload is the AWS service limit.
type Options struct {
	// Name is a required nonblank source label and must not contain secrets.
	Name string
	// Priority zero selects config.PriorityDiscoveredProfile.
	Priority int
	// Optional allows plan-level ErrNotFound suppression for missing secrets.
	Optional bool
	// SecretID is a required nonblank name/ARN of at most MaximumSecretIDBytes; IAM owns access.
	SecretID string
	// VersionID selects a provider version; it is passed unchanged without syntax validation.
	VersionID string
	// VersionStage selects a provider stage; AWSCURRENT is supplied only when both
	// selectors are empty.
	VersionStage string
	// MaximumBytes selects 1..MaximumSecretBytes, with zero choosing that maximum;
	// checked after SDK/copy allocation.
	MaximumBytes int
}

type secretsManagerSource struct {
	client       Client
	info         config.SourceInfo
	secretID     string
	versionID    string
	versionStage string
	maximumBytes int
	parseJSON    func([]byte, jsonsource.Options) (config.Source, error)
}

// New constructs a repeatable sensitive JSON source. The caller owns AWS SDK
// configuration, credential resolution, client retries, response cleanup and IAM
// policy. Nil/typed-nil clients and invalid options are rejected before provider
// I/O. Version selectors are passed unchanged; if both are empty AWSCURRENT is used.
// Each successful load requires exactly one nonempty bounded binary/string payload;
// plaintext persists in decoded values. No Close or explicit retry is introduced.
func New(client Client, options Options) (config.Source, error) {
	if nilLike(client) {
		return nil, ErrClientRequired
	}
	options, err := normalizeOptions(options)
	if err != nil {
		return nil, err
	}

	return &secretsManagerSource{
		client: client,
		info: config.SourceInfo{
			Name: options.Name, Priority: options.Priority,
			Sensitive: true, Optional: options.Optional,
		},
		secretID: options.SecretID, versionID: options.VersionID,
		versionStage: options.VersionStage, maximumBytes: options.MaximumBytes,
		parseJSON: jsonsource.Bytes,
	}, nil
}

func (s *secretsManagerSource) Info() config.SourceInfo { return s.info }

func (s *secretsManagerSource) Load(ctx context.Context) (config.Document, error) {
	if err := ctx.Err(); err != nil {
		return config.Document{}, err
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: stringPointer(s.secretID),
	}
	if s.versionID != "" {
		input.VersionId = stringPointer(s.versionID)
	}
	if s.versionStage != "" {
		input.VersionStage = stringPointer(s.versionStage)
	}
	output, err := s.client.GetSecretValue(ctx, input)
	if err != nil {
		var missing *types.ResourceNotFoundException
		if errors.As(err, &missing) {
			return config.Document{}, config.ErrNotFound
		}

		return config.Document{}, &operationError{cause: err}
	}

	payload, err := payload(output, s.maximumBytes)
	if err != nil {
		return config.Document{}, err
	}
	defer clear(payload)

	parser, err := s.parseJSON(payload, jsonsource.Options{
		Name: s.info.Name, Priority: s.info.Priority, Sensitive: true,
		Limits: jsonsource.Limits{MaxBytes: int64(s.maximumBytes)},
	})
	if err != nil {
		return config.Document{}, err
	}

	return parser.Load(ctx)
}

type operationError struct {
	cause error
}

// Error returns the stable operation classification without SDK text.
func (e *operationError) Error() string { return ErrOperation.Error() }

// Unwrap exposes the classification and raw SDK cause; callers must not log it.
func (e *operationError) Unwrap() error { return errors.Join(ErrOperation, e.cause) }

// Format keeps every fmt verb on the redacted operation classification.
func (e *operationError) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(e.Error()))
}

func normalizeOptions(options Options) (Options, error) {
	if strings.TrimSpace(options.Name) == "" ||
		strings.TrimSpace(options.SecretID) == "" ||
		len(options.SecretID) > MaximumSecretIDBytes ||
		options.MaximumBytes < 0 ||
		options.MaximumBytes > MaximumSecretBytes {
		return Options{}, ErrInvalidOptions
	}
	if options.Priority == 0 {
		options.Priority = config.PriorityDiscoveredProfile
	}
	if options.MaximumBytes == 0 {
		options.MaximumBytes = MaximumSecretBytes
	}
	if options.VersionID == "" && options.VersionStage == "" {
		options.VersionStage = "AWSCURRENT"
	}

	return options, nil
}

func payload(
	output *secretsmanager.GetSecretValueOutput,
	maximumBytes int,
) ([]byte, error) {
	if output == nil ||
		(output.SecretString == nil) == (output.SecretBinary == nil) {
		return nil, ErrInvalidResponse
	}

	// Admission follows SDK materialization and this local copy. Clearing it
	// cannot erase SDK strings/buffers or parser/decoded allocations.
	var value []byte
	if output.SecretString != nil {
		value = []byte(*output.SecretString)
	} else {
		value = append([]byte(nil), output.SecretBinary...)
	}
	if len(value) == 0 || len(value) > maximumBytes {
		clear(value)

		return nil, ErrInvalidResponse
	}

	return value, nil
}

func nilLike(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func stringPointer(value string) *string {
	return &value
}
