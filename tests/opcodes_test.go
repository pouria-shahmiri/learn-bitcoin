package tests

import (
	"testing"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
)

func TestOpcodeConstants(t *testing.T) {
	tests := []struct {
		opcode byte
		name   string
	}{
		{script.OP_0, "OP_0"},
		{script.OP_1, "OP_1"},
		{script.OP_16, "OP_16"},
		{script.OP_DUP, "OP_DUP"},
		{script.OP_HASH160, "OP_HASH160"},
		{script.OP_EQUALVERIFY, "OP_EQUALVERIFY"},
		{script.OP_CHECKSIG, "OP_CHECKSIG"},
	}

	for _, tt := range tests {
		name := script.OpcodeName(tt.opcode)
		if name != tt.name {
			t.Errorf("Opcode name mismatch: expected %s, got %s", tt.name, name)
		}
	}
}

func TestSmallIntOpcodes(t *testing.T) {
	// Test OP_1 through OP_16
	for i := 1; i <= 16; i++ {
		opcode := byte(script.OP_1 + i - 1)

		if !script.IsSmallInt(opcode) {
			t.Errorf("OP_%d not recognized as small int", i)
		}

		value := script.SmallIntValue(opcode)
		if value != i {
			t.Errorf("OP_%d: expected value %d, got %d", i, i, value)
		}
	}
}

func TestOpDup(t *testing.T) {
	scriptBytes := []byte{
		0x01, 0x42, // Push 0x42
		script.OP_DUP, // Duplicate it
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Fatalf("OP_DUP failed: %v", err)
	}

	if engine.Stack().Size() != 2 {
		t.Errorf("Expected 2 items on stack, got %d", engine.Stack().Size())
	}
}

func TestOpEqual(t *testing.T) {
	// Test equal values
	scriptBytes := []byte{
		0x02, 0x12, 0x34, // Push 0x1234
		0x02, 0x12, 0x34, // Push 0x1234
		script.OP_EQUAL, // Should push true (0x01)
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Fatalf("OP_EQUAL failed: %v", err)
	}

	top, _ := engine.Stack().Peek()
	if len(top) != 1 || top[0] != 0x01 {
		t.Error("OP_EQUAL should push 0x01 for equal values")
	}
}

func TestOpEqualVerify(t *testing.T) {
	// Test successful EQUALVERIFY
	scriptBytes := []byte{
		0x01, 0x42, // Push 0x42
		0x01, 0x42, // Push 0x42
		script.OP_EQUALVERIFY, // Should succeed and consume both
		script.OP_1,           // Push 1 to pass
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Errorf("OP_EQUALVERIFY should succeed: %v", err)
	}
}

func TestOpHash160(t *testing.T) {
	scriptBytes := []byte{
		0x04, 0x01, 0x02, 0x03, 0x04, // Push 4 bytes
		script.OP_HASH160, // Hash it
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Fatalf("OP_HASH160 failed: %v", err)
	}

	result, _ := engine.Stack().Peek()
	if len(result) != 20 {
		t.Errorf("OP_HASH160 should produce 20-byte hash, got %d", len(result))
	}
}

func TestOpVerifySuccess(t *testing.T) {
	scriptBytes := []byte{
		script.OP_1,      // Push true
		script.OP_VERIFY, // Should succeed
		script.OP_1,      // Push 1 to pass
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Errorf("OP_VERIFY should succeed with true value: %v", err)
	}
}

func TestOpVerifyFailure(t *testing.T) {
	scriptBytes := []byte{
		script.OP_0,      // Push false
		script.OP_VERIFY, // Should fail
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err == nil {
		t.Error("OP_VERIFY should fail with false value")
	}
}

func TestOpReturn(t *testing.T) {
	scriptBytes := []byte{
		script.OP_1,      // Push 1
		script.OP_RETURN, // Should fail immediately
	}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err == nil {
		t.Error("OP_RETURN should cause execution to fail")
	}
}
