package encoding

import (
	"crypto/sha256"
	"errors"
)

// EncodeBase58Check encodes data with version byte and checksum
// Format: [version][data][checksum(4 bytes)]
func EncodeBase58Check(version byte, data []byte) string {
	// Prepend version byte
	payload := make([]byte, 1+len(data))
	payload[0] = version
	copy(payload[1:], data)
	
	// Calculate checksum (first 4 bytes of double SHA-256)
	checksum := doubleSHA256(payload)[:4]
	
	// Append checksum
	fullPayload := append(payload, checksum...)
	
	return EncodeBase58(fullPayload)
}

// DecodeBase58Check decodes and verifies Base58Check encoded data
func DecodeBase58Check(input string) (version byte, data []byte, err error) {
	decoded, err := DecodeBase58(input)
	if err != nil {
		return 0, nil, err
	}
	
	if len(decoded) < 5 {
		return 0, nil, errors.New("decoded data too short")
	}
	
	// Split payload and checksum
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	
	// Verify checksum
	expectedChecksum := doubleSHA256(payload)[:4]
	for i := 0; i < 4; i++ {
		if checksum[i] != expectedChecksum[i] {
			return 0, nil, errors.New("checksum mismatch")
		}
	}
	
	// Extract version and data
	version = payload[0]
	data = payload[1:]
	
	return version, data, nil
}

// doubleSHA256 performs double SHA-256 hash
func doubleSHA256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}