package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ByteSize is a count of bytes decoded from values such as 10MiB or 2GB.
type ByteSize int64

const (
	// KiB is 1024 bytes.
	KiB ByteSize = 1 << 10
	// MiB is 1024 KiB.
	MiB ByteSize = 1 << 20
	// GiB is 1024 MiB.
	GiB ByteSize = 1 << 30
	// KB is 1000 bytes.
	KB ByteSize = 1_000
	// MB is 1000 KB.
	MB ByteSize = 1_000_000
	// GB is 1000 MB.
	GB ByteSize = 1_000_000_000
)

// UnmarshalText parses a non-negative decimal integer followed by optional B,
// KB/MB/GB or KiB/MiB/GiB (case-sensitive). Surrounding and pre-unit whitespace
// is allowed; fractions and overflowing products fail without changing b.
// A nil receiver is not supported.
func (b *ByteSize) UnmarshalText(text []byte) error {
	input := strings.TrimSpace(string(text))
	units := []struct {
		suffix string
		scale  ByteSize
	}{
		{"KiB", KiB}, {"MiB", MiB}, {"GiB", GiB},
		{"KB", KB}, {"MB", MB}, {"GB", GB}, {"B", 1},
	}
	for _, unit := range units {
		if strings.HasSuffix(input, unit.suffix) {
			number := strings.TrimSpace(strings.TrimSuffix(input, unit.suffix))
			value, err := strconv.ParseInt(number, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid byte size")
			}
			if value < 0 || value > math.MaxInt64/int64(unit.scale) {
				return fmt.Errorf("invalid byte size")
			}
			*b = ByteSize(value) * unit.scale
			return nil
		}
	}
	value, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid byte size")
	}
	if value < 0 {
		return fmt.Errorf("invalid byte size")
	}
	*b = ByteSize(value)
	return nil
}
