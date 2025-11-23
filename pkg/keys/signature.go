package keys

import (
	"encoding/hex"
	"fmt"
	
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Signature represents an ECDSA signature
type Signature struct {
	sig *ecdsa.Signature
}

// Serialize returns DER-encoded signature
func (s *Signature) Serialize() []byte {
	return s.sig.Serialize()
}

// String returns hex representation
func (s *Signature) String() string {
	return hex.EncodeToString(s.Serialize())
}

// ParseSignature parses a DER-encoded signature
func ParseSignature(data []byte) (*Signature, error) {
	sig, err := ecdsa.ParseDERSignature(data)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}
	
	return &Signature{sig: sig}, nil
}

// Compact returns compact signature format (64 bytes: 32-byte R + 32-byte S)
func (s *Signature) Compact() []byte {
	sig := make([]byte, 64)
	
	// Get R and S values as 32-byte arrays
	var rBytes, sBytes [32]byte
	r := s.sig.R()
	sVal := s.sig.S()
	r.PutBytes(&rBytes)
	sVal.PutBytes(&sBytes)
	
	copy(sig[0:32], rBytes[:])
	copy(sig[32:64], sBytes[:])
	
	return sig
}