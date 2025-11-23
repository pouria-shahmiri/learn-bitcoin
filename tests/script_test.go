package tests

import (
	"testing"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
)

func TestStackOperations(t *testing.T) {
	stack := script.NewStack()

	// Test push and pop
	stack.Push([]byte{0x01})
	stack.Push([]byte{0x02})

	if stack.Size() != 2 {
		t.Errorf("Expected size 2, got %d", stack.Size())
	}

	item, err := stack.Pop()
	if err != nil {
		t.Fatal(err)
	}
	if item[0] != 0x02 {
		t.Errorf("Expected 0x02, got %x", item)
	}
}

func TestStackDup(t *testing.T) {
	stack := script.NewStack()
	stack.Push([]byte{0x42})

	err := stack.Dup()
	if err != nil {
		t.Fatal(err)
	}

	if stack.Size() != 2 {
		t.Errorf("Expected size 2 after DUP, got %d", stack.Size())
	}

	item1, _ := stack.Pop()
	item2, _ := stack.Pop()

	if item1[0] != item2[0] {
		t.Error("DUP did not duplicate correctly")
	}
}

func TestStackSwap(t *testing.T) {
	stack := script.NewStack()
	stack.Push([]byte{0x01})
	stack.Push([]byte{0x02})

	err := stack.Swap()
	if err != nil {
		t.Fatal(err)
	}

	top, _ := stack.Pop()
	if top[0] != 0x01 {
		t.Errorf("Expected 0x01 on top after swap, got %x", top)
	}
}

func TestSimpleScriptExecution(t *testing.T) {
	// Script: Push 1
	scriptBytes := []byte{script.OP_1}

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Errorf("Simple script failed: %v", err)
	}

	if engine.Stack().Size() != 1 {
		t.Error("Expected 1 item on stack")
	}
}

func TestP2PKHScriptStructure(t *testing.T) {
	// Create a dummy pubkey hash
	pubKeyHash := make([]byte, 20)
	for i := range pubKeyHash {
		pubKeyHash[i] = byte(i)
	}

	// Generate P2PKH script
	lockingScript, err := script.P2PKH(pubKeyHash)
	if err != nil {
		t.Fatal(err)
	}

	// Verify structure
	if len(lockingScript) != 25 {
		t.Errorf("P2PKH script should be 25 bytes, got %d", len(lockingScript))
	}

	// Verify it's recognized as P2PKH
	if !script.IsP2PKH(lockingScript) {
		t.Error("Generated script not recognized as P2PKH")
	}

	// Extract and verify pubkey hash
	extracted, err := script.ExtractP2PKHAddress(lockingScript)
	if err != nil {
		t.Fatal(err)
	}

	for i := range pubKeyHash {
		if extracted[i] != pubKeyHash[i] {
			t.Error("Extracted pubkey hash doesn't match")
			break
		}
	}
}

func TestScriptBuilder(t *testing.T) {
	builder := script.NewBuilder()

	// Build: Push 1, Push 2, Drop, (leaves 1 on stack)
	builder.AddInt(1)
	builder.AddInt(2)
	builder.AddOp(script.OP_DROP)

	scriptBytes := builder.Script()

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		t.Errorf("Built script failed: %v", err)
	}

	// Check final stack
	if engine.Stack().Size() != 1 {
		t.Errorf("Expected 1 item on stack, got %d", engine.Stack().Size())
	}

	val, _ := engine.Stack().AsInt()
	if val != 1 {
		t.Errorf("Expected 1 on stack, got %d", val)
	}
}

func TestOpcodeNames(t *testing.T) {
	tests := []struct {
		opcode byte
		name   string
	}{
		{script.OP_DUP, "OP_DUP"},
		{script.OP_HASH160, "OP_HASH160"},
		{script.OP_CHECKSIG, "OP_CHECKSIG"},
		{script.OP_EQUAL, "OP_EQUAL"},
	}

	for _, tt := range tests {
		name := script.OpcodeName(tt.opcode)
		if name != tt.name {
			t.Errorf("Expected %s, got %s", tt.name, name)
		}
	}
}

func TestDisassemble(t *testing.T) {
	// Create simple script
	scriptBytes := []byte{script.OP_DUP, script.OP_HASH160, 0x01, 0x42, script.OP_EQUALVERIFY}

	asm := script.DisassembleScript(scriptBytes)

	// Should contain opcode names
	if len(asm) == 0 {
		t.Error("Disassembly returned empty string")
	}

	t.Logf("Disassembly: %s", asm)
}

func TestScriptNumEncoding(t *testing.T) {
	stack := script.NewStack()

	// Test various integers
	testCases := []int64{0, 1, -1, 127, -127, 128, -128, 32767}

	for _, n := range testCases {
		stack.PushInt(n)
		result, err := stack.AsInt()
		if err != nil {
			t.Errorf("Failed to convert %d: %v", n, err)
		}
		if result != n {
			t.Errorf("Expected %d, got %d", n, result)
		}
		stack.Pop() // Clean up
	}
}
