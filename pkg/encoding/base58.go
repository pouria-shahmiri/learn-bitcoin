package encoding

import (
	"errors"
	"math/big"
)

// Base58 alphabet (Bitcoin version - no 0, O, I, l)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	base58Base      = big.NewInt(58)
	bigZero         = big.NewInt(0)
	base58AlphabetMap [128]int8
)

func init() {
	// Build reverse lookup map
	for i := range base58AlphabetMap {
		base58AlphabetMap[i] = -1
	}
	for i, c := range base58Alphabet {
		base58AlphabetMap[c] = int8(i)
	}
}

// EncodeBase58 encodes bytes to Base58 string
func EncodeBase58(data []byte) string {
	// Convert bytes to big integer
	x := new(big.Int).SetBytes(data)
	
	// Encode to base58
	var result []byte
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
		x.DivMod(x, base58Base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}
	
	// Preserve leading zeros
	for _, b := range data {
		if b != 0 {
			break
		}
		result = append(result, base58Alphabet[0])
	}
	
	// Reverse result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	
	return string(result)
}

// DecodeBase58 decodes Base58 string to bytes
func DecodeBase58(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, nil
	}
	
	// Convert string to big integer
	x := big.NewInt(0)
	for _, c := range input {
		if c > 127 || base58AlphabetMap[c] == -1 {
			return nil, ErrInvalidBase58
		}
		x.Mul(x, base58Base)
		x.Add(x, big.NewInt(int64(base58AlphabetMap[c])))
	}
	
	// Convert to bytes
	decoded := x.Bytes()
	
	// Preserve leading '1's (zeros)
	for _, c := range input {
		if c != rune(base58Alphabet[0]) {
			break
		}
		decoded = append([]byte{0}, decoded...)
	}
	
	return decoded, nil
}

// ErrInvalidBase58 is returned for invalid Base58 strings
var ErrInvalidBase58 = errors.New("invalid base58 string")