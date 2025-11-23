package security

import (
	"bytes"
	"crypto/rand"
	"fmt"
)

// FuzzTester provides fuzzing capabilities for testing
type FuzzTester struct {
	minSize int
	maxSize int
}

// NewFuzzTester creates a new fuzz tester
func NewFuzzTester(minSize, maxSize int) *FuzzTester {
	return &FuzzTester{
		minSize: minSize,
		maxSize: maxSize,
	}
}

// GenerateRandomBytes generates random bytes for fuzzing
func (ft *FuzzTester) GenerateRandomBytes() ([]byte, error) {
	size := ft.minSize
	if ft.maxSize > ft.minSize {
		randSize := make([]byte, 1)
		if _, err := rand.Read(randSize); err != nil {
			return nil, err
		}
		size = ft.minSize + int(randSize[0])%(ft.maxSize-ft.minSize)
	}

	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return nil, err
	}

	return data, nil
}

// MutateBytes mutates existing bytes for fuzzing
func (ft *FuzzTester) MutateBytes(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	mutated := make([]byte, len(data))
	copy(mutated, data)

	// Random mutation strategies
	strategy := make([]byte, 1)
	rand.Read(strategy)

	switch strategy[0] % 5 {
	case 0: // Bit flip
		pos := make([]byte, 1)
		rand.Read(pos)
		idx := int(pos[0]) % len(mutated)
		mutated[idx] ^= 1 << (pos[0] % 8)

	case 1: // Byte flip
		pos := make([]byte, 1)
		rand.Read(pos)
		idx := int(pos[0]) % len(mutated)
		mutated[idx] ^= 0xFF

	case 2: // Insert random byte
		pos := make([]byte, 1)
		rand.Read(pos)
		idx := int(pos[0]) % len(mutated)
		randByte := make([]byte, 1)
		rand.Read(randByte)
		mutated[idx] = randByte[0]

	case 3: // Delete byte (if possible)
		if len(mutated) > 1 {
			pos := make([]byte, 1)
			rand.Read(pos)
			idx := int(pos[0]) % len(mutated)
			mutated = append(mutated[:idx], mutated[idx+1:]...)
		}

	case 4: // Duplicate byte
		pos := make([]byte, 1)
		rand.Read(pos)
		idx := int(pos[0]) % len(mutated)
		mutated = append(mutated[:idx+1], mutated[idx:]...)
	}

	return mutated
}

// GenerateMalformedMessage generates malformed Bitcoin messages
func (ft *FuzzTester) GenerateMalformedMessage() []byte {
	data, _ := ft.GenerateRandomBytes()
	return data
}

// TestMessageParsing tests message parsing with fuzz data
func (ft *FuzzTester) TestMessageParsing(parseFunc func([]byte) error, iterations int) []error {
	var errors []error

	for i := 0; i < iterations; i++ {
		data, err := ft.GenerateRandomBytes()
		if err != nil {
			errors = append(errors, fmt.Errorf("iteration %d: failed to generate data: %w", i, err))
			continue
		}

		err = parseFunc(data)
		if err != nil {
			// This is expected for fuzz testing - we're looking for panics/crashes
			// not parsing errors
		}
	}

	return errors
}

// InputValidator validates and sanitizes inputs
type InputValidator struct {
	maxSize int
}

// NewInputValidator creates an input validator
func NewInputValidator(maxSize int) *InputValidator {
	return &InputValidator{
		maxSize: maxSize,
	}
}

// ValidateMessageSize validates message size
func (iv *InputValidator) ValidateMessageSize(data []byte) error {
	if len(data) > iv.maxSize {
		return fmt.Errorf("message size %d exceeds maximum %d", len(data), iv.maxSize)
	}
	return nil
}

// ValidateNoNullBytes checks for null bytes in string data
func (iv *InputValidator) ValidateNoNullBytes(data []byte) error {
	if bytes.Contains(data, []byte{0}) {
		return fmt.Errorf("data contains null bytes")
	}
	return nil
}

// SanitizeString removes potentially dangerous characters
func (iv *InputValidator) SanitizeString(s string) string {
	// Remove null bytes and control characters
	var result []byte
	for i := 0; i < len(s); i++ {
		if s[i] >= 32 && s[i] <= 126 {
			result = append(result, s[i])
		}
	}
	return string(result)
}

// ValidateIPAddress validates IP address format
func (iv *InputValidator) ValidateIPAddress(ip string) error {
	// Basic validation - in production use net.ParseIP
	if len(ip) < 7 || len(ip) > 45 {
		return fmt.Errorf("invalid IP address length")
	}
	return nil
}

// ValidatePort validates port number
func (iv *InputValidator) ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}
	return nil
}
