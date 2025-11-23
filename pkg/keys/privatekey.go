package keys

import (
	"fmt"
	
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/encoding"
)

// PrivateKey represents a Bitcoin private key
type PrivateKey struct {
	key *secp256k1.PrivateKey
}

// GeneratePrivateKey generates a new random private key
func GeneratePrivateKey() (*PrivateKey, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	
	return &PrivateKey{key: key}, nil
}

// NewPrivateKeyFromBytes creates private key from bytes
func NewPrivateKeyFromBytes(data []byte) (*PrivateKey, error) {
	if len(data) != 32 {
		return nil, fmt.Errorf("private key must be 32 bytes, got %d", len(data))
	}
	
	key := secp256k1.PrivKeyFromBytes(data)
	return &PrivateKey{key: key}, nil
}

// Bytes returns the private key as 32 bytes
func (pk *PrivateKey) Bytes() []byte {
	return pk.key.Serialize()
}

// PublicKey derives the public key from private key
func (pk *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{
		key: pk.key.PubKey(),
	}
}

// ToWIF exports private key in Wallet Import Format (WIF)
// WIF format: Base58Check([version][32-byte key][compression flag])
func (pk *PrivateKey) ToWIF(compressed bool) string {
	// Mainnet private key version byte is 0x80
	version := byte(0x80)
	
	data := pk.Bytes()
	
	// Add compression flag if compressed
	if compressed {
		data = append(data, 0x01)
	}
	
	return encoding.EncodeBase58Check(version, data)
}

// FromWIF imports private key from WIF format
func FromWIF(wif string) (*PrivateKey, bool, error) {
	version, data, err := encoding.DecodeBase58Check(wif)
	if err != nil {
		return nil, false, fmt.Errorf("invalid WIF: %w", err)
	}
	
	// Check version byte (0x80 for mainnet)
	if version != 0x80 {
		return nil, false, fmt.Errorf("invalid WIF version: %x", version)
	}
	
	// Check length and compression flag
	compressed := false
	if len(data) == 33 && data[32] == 0x01 {
		compressed = true
		data = data[:32]
	} else if len(data) != 32 {
		return nil, false, fmt.Errorf("invalid WIF key length: %d", len(data))
	}
	
	key, err := NewPrivateKeyFromBytes(data)
	if err != nil {
		return nil, false, err
	}
	
	return key, compressed, nil
}

// Sign signs a message hash with the private key
func (pk *PrivateKey) Sign(hash []byte) (*Signature, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	
	sig := ecdsa.Sign(pk.key, hash)
	
	return &Signature{sig: sig}, nil
}

// String returns hex representation (for debugging only - never expose!)
func (pk *PrivateKey) String() string {
	return fmt.Sprintf("%x", pk.Bytes())
}