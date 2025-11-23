package tests

import (
	"testing"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/transaction"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

func TestValidateTransactionBasic(t *testing.T) {
	// Create a valid transaction
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{1, 2, 3},
				OutputIndex:     0,
				SignatureScript: []byte{0x01, 0x02},
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{
				Value:        100000000,
				PubKeyScript: []byte{0x76, 0xa9},
			},
		},
		LockTime: 0,
	}

	err := transaction.ValidateTransaction(tx)
	if err != nil {
		t.Errorf("Valid transaction failed validation: %v", err)
	}
}

func TestValidateTransactionNoInputs(t *testing.T) {
	tx := &types.Transaction{
		Version:  1,
		Inputs:   []types.TxInput{},
		Outputs:  []types.TxOutput{{Value: 100}},
		LockTime: 0,
	}

	err := transaction.ValidateTransaction(tx)
	if err == nil {
		t.Error("Transaction with no inputs should fail validation")
	}
}

func TestValidateTransactionNoOutputs(t *testing.T) {
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{PrevTxHash: types.Hash{1}},
		},
		Outputs:  []types.TxOutput{},
		LockTime: 0,
	}

	err := transaction.ValidateTransaction(tx)
	if err == nil {
		t.Error("Transaction with no outputs should fail validation")
	}
}

func TestValidateTransactionNegativeOutput(t *testing.T) {
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{PrevTxHash: types.Hash{1}},
		},
		Outputs: []types.TxOutput{
			{Value: -100},
		},
		LockTime: 0,
	}

	err := transaction.ValidateTransaction(tx)
	if err == nil {
		t.Error("Transaction with negative output should fail validation")
	}
}

func TestIsCoinbase(t *testing.T) {
	// Create coinbase transaction
	coinbase := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{}, // All zeros
				OutputIndex:     0xFFFFFFFF,
				SignatureScript: []byte{0x01, 0x02, 0x03},
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{Value: 5000000000},
		},
		LockTime: 0,
	}

	if !transaction.IsCoinbase(coinbase) {
		t.Error("Valid coinbase not recognized")
	}

	// Create non-coinbase transaction
	regular := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:  types.Hash{1, 2, 3},
				OutputIndex: 0,
			},
		},
		Outputs: []types.TxOutput{
			{Value: 100},
		},
		LockTime: 0,
	}

	if transaction.IsCoinbase(regular) {
		t.Error("Regular transaction incorrectly identified as coinbase")
	}
}

func TestValidateCoinbase(t *testing.T) {
	coinbase := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{
				PrevTxHash:      types.Hash{},
				OutputIndex:     0xFFFFFFFF,
				SignatureScript: []byte{0x03, 0x01, 0x02, 0x03}, // 4 bytes (valid length)
				Sequence:        0xFFFFFFFF,
			},
		},
		Outputs: []types.TxOutput{
			{Value: 5000000000},
		},
		LockTime: 0,
	}

	err := transaction.ValidateCoinbase(coinbase, 1)
	if err != nil {
		t.Errorf("Valid coinbase failed validation: %v", err)
	}
}

func TestCalculateFee(t *testing.T) {
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{PrevTxHash: types.Hash{1}, OutputIndex: 0},
			{PrevTxHash: types.Hash{2}, OutputIndex: 1},
		},
		Outputs: []types.TxOutput{
			{Value: 100000000},
			{Value: 50000000},
		},
		LockTime: 0,
	}

	prevOutputs := []types.TxOutput{
		{Value: 100000000},
		{Value: 60000000},
	}

	fee, err := transaction.CalculateFee(tx, prevOutputs)
	if err != nil {
		t.Fatalf("Failed to calculate fee: %v", err)
	}

	expectedFee := int64(10000000) // 160M input - 150M output = 10M fee
	if fee != expectedFee {
		t.Errorf("Expected fee %d, got %d", expectedFee, fee)
	}
}

func TestCalculateFeeNegative(t *testing.T) {
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{PrevTxHash: types.Hash{1}, OutputIndex: 0},
		},
		Outputs: []types.TxOutput{
			{Value: 200000000}, // Output exceeds input
		},
		LockTime: 0,
	}

	prevOutputs := []types.TxOutput{
		{Value: 100000000},
	}

	_, err := transaction.CalculateFee(tx, prevOutputs)
	if err == nil {
		t.Error("Should fail when outputs exceed inputs")
	}
}

func TestDuplicateInputs(t *testing.T) {
	// Transaction with duplicate inputs (double-spend within same tx)
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{PrevTxHash: types.Hash{1}, OutputIndex: 0},
			{PrevTxHash: types.Hash{1}, OutputIndex: 0}, // Duplicate!
		},
		Outputs: []types.TxOutput{
			{Value: 100},
		},
		LockTime: 0,
	}

	err := transaction.ValidateTransaction(tx)
	if err == nil {
		t.Error("Transaction with duplicate inputs should fail validation")
	}
}

func TestTransactionBuilder(t *testing.T) {
	builder := transaction.NewTxBuilder()

	// Add input
	builder.AddInput(types.Hash{1, 2, 3}, 0)

	// Add output
	builder.AddOutput(100000000, []byte{0x76, 0xa9})

	// Build transaction
	tx, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build transaction: %v", err)
	}

	// Verify structure
	if len(tx.Inputs) != 1 {
		t.Errorf("Expected 1 input, got %d", len(tx.Inputs))
	}

	if len(tx.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(tx.Outputs))
	}

	if tx.Outputs[0].Value != 100000000 {
		t.Errorf("Expected output value 100000000, got %d", tx.Outputs[0].Value)
	}
}

func TestCreateCoinbase(t *testing.T) {
	// Generate a test address (using mock for simplicity)
	testAddress := "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

	coinbase, err := transaction.CreateCoinbase(
		100,        // Block height
		5000000000, // 50 BTC reward
		testAddress,
		[]byte("Hello Bitcoin!"), // Extra data
	)

	if err != nil {
		t.Fatalf("Failed to create coinbase: %v", err)
	}

	// Verify it's a valid coinbase
	if !transaction.IsCoinbase(coinbase) {
		t.Error("Created transaction is not recognized as coinbase")
	}

	// Verify structure
	if len(coinbase.Inputs) != 1 {
		t.Error("Coinbase should have exactly 1 input")
	}

	if len(coinbase.Outputs) < 1 {
		t.Error("Coinbase should have at least 1 output")
	}

	if coinbase.Outputs[0].Value != 5000000000 {
		t.Errorf("Expected reward 5000000000, got %d", coinbase.Outputs[0].Value)
	}
}

func TestSigHashTypes(t *testing.T) {
	tests := []struct {
		hashType transaction.SigHashType
		expected string
	}{
		{transaction.SigHashAll, "ALL"},
		{transaction.SigHashNone, "NONE"},
		{transaction.SigHashSingle, "SINGLE"},
		{transaction.SigHashAll | transaction.SigHashAnyOneCanPay, "ALL|ANYONECANPAY"},
	}

	for _, tt := range tests {
		info := transaction.SignatureHashInfo(tt.hashType)
		if info != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, info)
		}
	}
}

func TestCalculateTransactionSize(t *testing.T) {
	// Test size estimation
	size := transaction.CalculateSize(2, 2)

	// Should be roughly: 4 (version) + 1 (input count) + 296 (2 inputs * 148) +
	//                    1 (output count) + 68 (2 outputs * 34) + 4 (locktime) = 374
	if size < 300 || size > 400 {
		t.Errorf("Size estimation seems off: %d bytes", size)
	}
}

func TestEstimateFee(t *testing.T) {
	feePerByte := int64(10) // 10 satoshis per byte

	fee := transaction.EstimateFee(1, 2, feePerByte)

	// For 1 input, 2 outputs, size ~226 bytes
	// Fee should be around 2260 satoshis
	if fee < 2000 || fee > 3000 {
		t.Errorf("Fee estimation seems off: %d satoshis", fee)
	}
}
