package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
)

func main() {
	fmt.Println("=== Bitcoin Learning - Milestone 3 ===")
	fmt.Println("Cryptography & Key Handling\n")

	// Demo 1: Generate keys
	demoKeyGeneration()

	// Demo 2: Bitcoin addresses
	demoAddressGeneration()

	// Demo 3: WIF import/export
	demoWIFFormat()

	// Demo 4: Sign and verify
	demoSignAndVerify()

	// Demo 5: Multiple addresses from same key
	demoMultipleFormats()

	// Demo 6: Address validation
	demoAddressValidation()

	fmt.Println("\n=== All demos completed successfully! ===")
}

func demoKeyGeneration() {
	fmt.Println("--- Demo 1: Key Generation ---")

	// Generate random private key
	privKey, err := keys.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}

	// Derive public key
	pubKey := privKey.PublicKey()

	fmt.Printf("Private Key (32 bytes): %s\n", privKey.String())
	fmt.Printf("Public Key (compressed): %s\n", pubKey.String())
	fmt.Printf("Public Key bytes: %d\n", len(pubKey.Bytes(true)))
	fmt.Println()
}

func demoAddressGeneration() {
	fmt.Println("--- Demo 2: Bitcoin Address Generation ---")

	// Generate key pair
	privKey, _ := keys.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	// Generate mainnet address
	address := pubKey.P2PKHAddress()
	fmt.Printf("Bitcoin Address (Mainnet): %s\n", address)
	fmt.Printf("  Starts with '1' (P2PKH)\n")

	// Generate testnet address
	testnetAddr := pubKey.TestnetP2PKHAddress()
	fmt.Printf("Bitcoin Address (Testnet): %s\n", testnetAddr)
	fmt.Printf("  Starts with 'm' or 'n' (Testnet P2PKH)\n")

	// Show hash160
	hash160 := pubKey.Hash160()
	fmt.Printf("Hash160 (RIPEMD160(SHA256(pubkey))): %x\n", hash160)
	fmt.Printf("Hash160 length: %d bytes\n", len(hash160))
	fmt.Println()
}

func demoWIFFormat() {
	fmt.Println("--- Demo 3: WIF (Wallet Import Format) ---")

	// Generate key
	privKey, _ := keys.GeneratePrivateKey()

	// Export to WIF (compressed)
	wifCompressed := privKey.ToWIF(true)
	fmt.Printf("WIF (compressed): %s\n", wifCompressed)

	// Export to WIF (uncompressed)
	wifUncompressed := privKey.ToWIF(false)
	fmt.Printf("WIF (uncompressed): %s\n", wifUncompressed)

	// Import from WIF
	importedKey, compressed, err := keys.FromWIF(wifCompressed)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nImported from WIF:\n")
	fmt.Printf("  Compressed: %v\n", compressed)
	fmt.Printf("  Keys match: %v\n", importedKey.String() == privKey.String())
	fmt.Println()
}

func demoSignAndVerify() {
	fmt.Println("--- Demo 4: Sign and Verify ---")

	// Generate key pair
	privKey, _ := keys.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	// Message to sign
	message := []byte("Hello, Bitcoin!")
	messageHash := sha256.Sum256(message)

	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Message Hash: %x\n", messageHash)

	// Sign message
	signature, err := privKey.Sign(messageHash[:])
	if err != nil {
		panic(err)
	}

	fmt.Printf("Signature (DER): %s\n", signature.String())
	fmt.Printf("Signature length: %d bytes\n", len(signature.Serialize()))

	// Verify signature
	valid := pubKey.Verify(messageHash[:], signature)
	fmt.Printf("\nSignature valid: %v ✓\n", valid)

	// Try with wrong message
	wrongHash := sha256.Sum256([]byte("Wrong message"))
	invalidSig := pubKey.Verify(wrongHash[:], signature)
	fmt.Printf("Wrong message valid: %v ✓\n", invalidSig)
	fmt.Println()
}

func demoMultipleFormats() {
	fmt.Println("--- Demo 5: Multiple Address Formats ---")

	// Same private key
	privKey, _ := keys.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	fmt.Println("Same key, different representations:")
	fmt.Printf("  Private (hex): %s\n", privKey.String())
	fmt.Printf("  Private (WIF): %s\n", privKey.ToWIF(true))
	fmt.Printf("  Public (hex):  %s\n", pubKey.String())
	fmt.Printf("  Address:       %s\n", pubKey.P2PKHAddress())
	fmt.Println()
}

func demoAddressValidation() {
	fmt.Println("--- Demo 6: Address Validation ---")

	// Generate valid address
	privKey, _ := keys.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	validAddress := pubKey.P2PKHAddress()

	fmt.Printf("Valid address: %s\n", validAddress)

	// Decode and validate
	decoded, err := keys.DecodeAddress(validAddress)
	if err != nil {
		panic(err)
	}

	fmt.Printf("  Version byte: 0x%02x\n", decoded.Version())
	fmt.Printf("  Is P2PKH: %v\n", decoded.IsP2PKH())
	fmt.Printf("  Hash: %x\n", decoded.Hash())

	// Try invalid address (wrong checksum)
	invalidAddress := "1InvalidAddress123"
	_, err = keys.DecodeAddress(invalidAddress)
	if err != nil {
		fmt.Printf("\nInvalid address rejected: %v ✓\n", err)
	}
	fmt.Println()
}
