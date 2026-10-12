package config

import (
	"encoding/json"
	"fmt"
	"io"
)

// Redacted is the stable replacement used for secret values.
const Redacted = "[REDACTED]"

// Secret stores sensitive text with redacted default formatting and marshaling.
// Reveal is the supported plaintext accessor, not a security or protected-memory
// boundary. The string is retained/shared, is not encrypted and cannot be erased
// reliably in Go. The zero Secret holds empty text; equality compares its text.
type Secret struct {
	value string
}

// NewSecret retains sensitive text without copying or validating it.
func NewSecret(value string) Secret { return Secret{value: value} }

// Reveal explicitly returns the sensitive text.
func (s Secret) Reveal() string { return s.value }

// String returns Redacted, including for an empty Secret.
func (Secret) String() string { return Redacted }

// GoString returns Redacted for Go-syntax diagnostics.
func (Secret) GoString() string { return Redacted }

// Format redacts every fmt formatting verb.
func (Secret) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, Redacted)
}

// MarshalText prevents text-based encoders from exposing the value.
func (Secret) MarshalText() ([]byte, error) { return []byte(Redacted), nil }

// MarshalJSON prevents JSON diagnostics from exposing the value.
func (Secret) MarshalJSON() ([]byte, error) { return json.Marshal(Redacted) }

// UnmarshalText allows strict decoders to populate Secret from string values.
func (s *Secret) UnmarshalText(text []byte) error {
	s.value = string(text)
	return nil
}
