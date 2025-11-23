package script

import (
	"fmt"
)

// P2PKH creates a Pay-to-PubKey-Hash locking script
// Format: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
func P2PKH(pubKeyHash []byte) ([]byte, error) {
	if len(pubKeyHash) != 20 {
		return nil, fmt.Errorf("pubKeyHash must be 20 bytes, got %d", len(pubKeyHash))
	}

	script := []byte{
		OP_DUP,
		OP_HASH160,
		byte(len(pubKeyHash)), // Push 20 bytes
	}
	script = append(script, pubKeyHash...)
	script = append(script, OP_EQUALVERIFY, OP_CHECKSIG)

	return script, nil
}

// P2PKHUnlockingScript creates an unlocking script for P2PKH
// Format: <signature> <pubKey>
func P2PKHUnlockingScript(signature, pubKey []byte) []byte {
	var script []byte

	// Push signature
	script = append(script, byte(len(signature)))
	script = append(script, signature...)

	// Push public key
	script = append(script, byte(len(pubKey)))
	script = append(script, pubKey...)

	return script
}

// IsP2PKH checks if script is a P2PKH locking script
func IsP2PKH(script []byte) bool {
	if len(script) != 25 {
		return false
	}

	return script[0] == OP_DUP &&
		script[1] == OP_HASH160 &&
		script[2] == 20 && // Push 20 bytes
		script[23] == OP_EQUALVERIFY &&
		script[24] == OP_CHECKSIG
}

// ExtractP2PKHAddress extracts the pubkey hash from a P2PKH script
func ExtractP2PKHAddress(script []byte) ([]byte, error) {
	if !IsP2PKH(script) {
		return nil, fmt.Errorf("not a P2PKH script")
	}

	// Extract bytes 3-22 (the pubkey hash)
	return script[3:23], nil
}

// ExecuteP2PKH executes a complete P2PKH transaction
func ExecuteP2PKH(unlocking, locking []byte) error {
	// Combine scripts: unlocking + locking
	combined := append(unlocking, locking...)

	engine := NewEngine(combined)
	return engine.Execute()
}

// DisassembleScript converts script bytes to human-readable format
func DisassembleScript(script []byte) string {
	var result string
	pc := 0

	for pc < len(script) {
		opcode := script[pc]
		pc++

		// Handle data push
		if opcode > 0 && opcode <= 0x4b {
			if pc+int(opcode) > len(script) {
				result += fmt.Sprintf("[INVALID PUSH %d] ", opcode)
				break
			}
			data := script[pc : pc+int(opcode)]
			pc += int(opcode)
			result += fmt.Sprintf("[%x] ", data)
		} else {
			result += OpcodeName(opcode) + " "
		}
	}

	return result
}
