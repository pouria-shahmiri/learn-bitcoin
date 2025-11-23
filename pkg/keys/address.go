package keys

import (
	"fmt"
	
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/encoding"
)

// Address types
const (
	// P2PKH (Pay-to-PubKey-Hash) - starts with '1'
	AddressTypeP2PKH byte = 0x00
	
	// P2SH (Pay-to-Script-Hash) - starts with '3'
	AddressTypeP2SH byte = 0x05
	
	// Testnet P2PKH - starts with 'm' or 'n'
	AddressTypeTestnetP2PKH byte = 0x6f
	
	// Testnet P2SH - starts with '2'
	AddressTypeTestnetP2SH byte = 0xc4
)

// Address represents a Bitcoin address
type Address struct {
	version byte
	hash    []byte // 20-byte hash
}

// NewAddress creates an address from version and hash
func NewAddress(version byte, hash []byte) (*Address, error) {
	if len(hash) != 20 {
		return nil, fmt.Errorf("hash must be 20 bytes, got %d", len(hash))
	}
	
	return &Address{
		version: version,
		hash:    make([]byte, 20),
	}, nil
}

// P2PKHAddress creates a Pay-to-PubKey-Hash address
func (pub *PublicKey) P2PKHAddress() string {
	hash160 := pub.Hash160()
	return encoding.EncodeBase58Check(AddressTypeP2PKH, hash160)
}

// TestnetP2PKHAddress creates a testnet P2PKH address
func (pub *PublicKey) TestnetP2PKHAddress() string {
	hash160 := pub.Hash160()
	return encoding.EncodeBase58Check(AddressTypeTestnetP2PKH, hash160)
}

// DecodeAddress decodes a Bitcoin address
func DecodeAddress(address string) (*Address, error) {
	version, hash, err := encoding.DecodeBase58Check(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	
	if len(hash) != 20 {
		return nil, fmt.Errorf("invalid address hash length: %d", len(hash))
	}
	
	addr := &Address{
		version: version,
		hash:    hash,
	}
	
	return addr, nil
}

// String returns the Base58Check encoded address
func (addr *Address) String() string {
	return encoding.EncodeBase58Check(addr.version, addr.hash)
}

// IsP2PKH checks if address is Pay-to-PubKey-Hash
func (addr *Address) IsP2PKH() bool {
	return addr.version == AddressTypeP2PKH || 
	       addr.version == AddressTypeTestnetP2PKH
}

// IsP2SH checks if address is Pay-to-Script-Hash
func (addr *Address) IsP2SH() bool {
	return addr.version == AddressTypeP2SH || 
	       addr.version == AddressTypeTestnetP2SH
}

// Hash returns the 20-byte address hash
func (addr *Address) Hash() []byte {
	return addr.hash
}

// Version returns the address version byte
func (addr *Address) Version() byte {
	return addr.version
}