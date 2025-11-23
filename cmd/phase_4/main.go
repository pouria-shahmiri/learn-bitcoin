package main

import (
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/keys"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/script"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/serialization"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/transaction"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

func main() {
	fmt.Println("=== Bitcoin Learning - Milestone 4 ===")
	fmt.Println("Transactions & Script VM\n")

	// Demo 1: Basic stack operations
	demoStackOperations()

	// Demo 2: Simple script execution
	demoSimpleScript()

	// Demo 3: P2PKH script generation
	demoP2PKHScript()

	// Demo 4: Script verification
	demoScriptVerification()

	// Demo 5: Script builder
	demoScriptBuilder()

	// Demo 6: Disassemble scripts
	demoDisassemble()

	// Demo 7: Transaction builder
	demoTransactionBuilder()

	// Demo 8: Coinbase transaction
	demoCoinbaseTransaction()

	// Demo 9: Fee calculation
	demoFeeCalculation()

	// Demo 10: Signature hash types
	demoSignatureHashTypes()

	fmt.Println("\n=== All demos completed successfully! ===")
}

func demoStackOperations() {
	fmt.Println("--- Demo 1: Stack Operations ---")

	stack := script.NewStack()

	// Push items
	stack.Push([]byte{0x01, 0x02})
	stack.Push([]byte{0x03, 0x04})
	stack.Push([]byte{0x05, 0x06})

	fmt.Printf("Stack after 3 pushes: %s\n", stack.String())
	fmt.Printf("Stack size: %d\n", stack.Size())

	// Pop item
	item, _ := stack.Pop()
	fmt.Printf("Popped: %x\n", item)
	fmt.Printf("Stack after pop: %s\n", stack.String())

	// Duplicate top
	stack.Dup()
	fmt.Printf("Stack after DUP: %s\n", stack.String())

	// Swap top two
	stack.Swap()
	fmt.Printf("Stack after SWAP: %s\n", stack.String())

	// Push integers
	stack.Clear()
	stack.PushInt(42)
	stack.PushInt(100)
	fmt.Printf("Stack with integers: %s\n", stack.String())

	val, _ := stack.AsInt()
	fmt.Printf("Top value as int: %d\n", val)

	fmt.Println()
}

func demoSimpleScript() {
	fmt.Println("--- Demo 2: Simple Script Execution ---")

	// Simple script: Push 1, Push 2, Add (if we had OP_ADD)
	// For now: Push data and verify it's on stack

	builder := script.NewBuilder()
	builder.AddInt(5)
	builder.AddInt(10)
	builder.AddOp(script.OP_DROP) // Drop the 10

	scriptBytes := builder.Script()
	fmt.Printf("Script bytes: %x\n", scriptBytes)
	fmt.Printf("Script ASM: %s\n", script.DisassembleScript(scriptBytes))

	engine := script.NewEngine(scriptBytes)
	err := engine.Execute()

	if err != nil {
		fmt.Printf("Execution failed: %v\n", err)
	} else {
		fmt.Println("Script executed successfully ✓")
		fmt.Printf("Final stack: %s\n", engine.Stack().String())
	}

	fmt.Println()
}

func demoP2PKHScript() {
	fmt.Println("--- Demo 3: P2PKH Script Generation ---")

	// Generate key pair
	privKey, _ := keys.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	// Get pubkey hash
	pubKeyHash := pubKey.Hash160()
	fmt.Printf("Public Key Hash: %x\n", pubKeyHash)

	// Create P2PKH locking script
	lockingScript, err := script.P2PKH(pubKeyHash)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Locking Script (hex): %x\n", lockingScript)
	fmt.Printf("Locking Script (asm): %s\n", script.DisassembleScript(lockingScript))
	fmt.Printf("Locking Script length: %d bytes\n", len(lockingScript))

	// Verify it's valid P2PKH
	isP2PKH := script.IsP2PKH(lockingScript)
	fmt.Printf("Is valid P2PKH: %v ✓\n", isP2PKH)

	// Extract address from script
	extractedHash, _ := script.ExtractP2PKHAddress(lockingScript)
	fmt.Printf("Extracted hash matches: %v ✓\n", string(extractedHash) == string(pubKeyHash))

	// Show address
	address := pubKey.P2PKHAddress()
	fmt.Printf("Bitcoin Address: %s\n", address)

	fmt.Println()
}

func demoScriptVerification() {
	fmt.Println("--- Demo 4: Script Verification Flow ---")

	// Generate key pair
	privKey, _ := keys.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	pubKeyHash := pubKey.Hash160()

	// Create locking script (what's in the output)
	lockingScript, _ := script.P2PKH(pubKeyHash)
	fmt.Println("Step 1: Locking script created (scriptPubKey)")
	fmt.Printf("  %s\n", script.DisassembleScript(lockingScript))

	// Create unlocking script (what's in the input)
	// For demo, we'll use dummy signature
	dummySig := []byte{0x30, 0x44, 0x02, 0x20}       // Dummy DER signature prefix
	dummySig = append(dummySig, make([]byte, 32)...) // 32-byte R
	dummySig = append(dummySig, 0x02, 0x20)          // S marker
	dummySig = append(dummySig, make([]byte, 32)...) // 32-byte S
	dummySig = append(dummySig, 0x01)                // SIGHASH_ALL

	pubKeyBytes := pubKey.Bytes(true)
	unlockingScript := script.P2PKHUnlockingScript(dummySig, pubKeyBytes)

	fmt.Println("\nStep 2: Unlocking script created (scriptSig)")
	fmt.Printf("  %s\n", script.DisassembleScript(unlockingScript))

	fmt.Println("\nStep 3: Execution flow:")
	fmt.Println("  Initial stack: []")
	fmt.Println("  Execute scriptSig:")
	fmt.Println("    - Push signature → [sig]")
	fmt.Println("    - Push public key → [sig, pubkey]")
	fmt.Println("  Execute scriptPubKey:")
	fmt.Println("    - OP_DUP → [sig, pubkey, pubkey]")
	fmt.Println("    - OP_HASH160 → [sig, pubkey, hash(pubkey)]")
	fmt.Println("    - Push expected hash → [sig, pubkey, hash(pubkey), expected_hash]")
	fmt.Println("    - OP_EQUALVERIFY → [sig, pubkey] (or fail)")
	fmt.Println("    - OP_CHECKSIG → [true/false]")

	// Note: Full execution requires signature verification
	// which we'll implement in the next section
	fmt.Println("\n✓ Script structure is valid")
	fmt.Println()
}

func demoScriptBuilder() {
	fmt.Println("--- Demo 5: Script Builder ---")

	// Build a custom script
	builder := script.NewBuilder()

	// Push some data
	builder.AddData([]byte("Hello"))
	builder.AddData([]byte("Bitcoin"))

	// Add some operations
	builder.AddOp(script.OP_DROP)
	builder.AddOp(script.OP_DROP)

	// Push true
	builder.AddInt(1)

	customScript := builder.Script()

	fmt.Printf("Custom script (hex): %x\n", customScript)
	fmt.Printf("Custom script (asm): %s\n", script.DisassembleScript(customScript))
	fmt.Printf("Custom script length: %d bytes\n", len(customScript))

	// Execute it
	engine := script.NewEngine(customScript)
	err := engine.Execute()

	if err != nil {
		fmt.Printf("Execution failed: %v\n", err)
	} else {
		fmt.Println("Custom script executed successfully ✓")
		fmt.Printf("Final stack: %s\n", engine.Stack().String())
	}

	// Another example: Math operations (simulated)
	fmt.Println("\nAnother example - Push 3 numbers:")
	builder2 := script.NewBuilder()
	builder2.AddInt(10)
	builder2.AddInt(20)
	builder2.AddInt(30)
	builder2.AddOp(script.OP_DROP) // Drop 30
	builder2.AddOp(script.OP_DROP) // Drop 20
	// Leaves 10 on stack

	script2 := builder2.Script()
	fmt.Printf("Script: %s\n", script.DisassembleScript(script2))

	engine2 := script.NewEngine(script2)
	engine2.Execute()
	fmt.Printf("Final stack: %s\n", engine2.Stack().String())

	fmt.Println()
}

func demoDisassemble() {
	fmt.Println("--- Demo 6: Disassemble Scripts ---")

	// Create various script types
	scripts := map[string][]byte{
		"Simple Push":       {0x01, 0xff}, // Push 1 byte (0xff)
		"OP_DUP OP_HASH160": {script.OP_DUP, script.OP_HASH160},
		"Push 1,2,3":        {script.OP_1, script.OP_2, script.OP_3},
		"OP_RETURN":         {script.OP_RETURN, 0x04, 't', 'e', 's', 't'},
		"Empty Script":      {},
	}

	for name, scriptBytes := range scripts {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  Hex: %x\n", scriptBytes)
		fmt.Printf("  ASM: %s\n", script.DisassembleScript(scriptBytes))
	}

	// Example: Disassemble a real P2PKH script
	fmt.Println("\nReal P2PKH Script:")
	privKey, _ := keys.GeneratePrivateKey()
	pubKeyHash := privKey.PublicKey().Hash160()
	p2pkhScript, _ := script.P2PKH(pubKeyHash)
	fmt.Printf("  Hex: %x\n", p2pkhScript)
	fmt.Printf("  ASM: %s\n", script.DisassembleScript(p2pkhScript))

	fmt.Println()
}

func demoTransactionBuilder() {
	fmt.Println("--- Demo 7: Transaction Builder ---")

	// Generate sender and receiver keys
	senderKey, _ := keys.GeneratePrivateKey()
	receiverKey, _ := keys.GeneratePrivateKey()

	senderAddress := senderKey.PublicKey().P2PKHAddress()
	receiverAddress := receiverKey.PublicKey().P2PKHAddress()

	fmt.Printf("Scenario: Alice sends Bitcoin to Bob\n")
	fmt.Printf("Alice (sender): %s\n", senderAddress)
	fmt.Printf("Bob (receiver): %s\n", receiverAddress)

	// Build transaction
	builder := transaction.NewTxBuilder()

	// Add input (spending from a previous output)
	// In reality, this would reference a real UTXO
	prevTxHash := crypto.DoubleSHA256([]byte("previous transaction"))
	builder.AddInput(prevTxHash, 0)

	// Add outputs
	builder.AddP2PKHOutput(100000000, receiverAddress) // Send 1 BTC to Bob
	builder.AddP2PKHOutput(40000000, senderAddress)    // Change: 0.4 BTC back to Alice
	// Implied fee: 0.1 BTC (if input was 1.5 BTC)

	// Build unsigned transaction
	tx, err := builder.Build()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nTransaction built:\n")
	fmt.Printf("  Version: %d\n", tx.Version)
	fmt.Printf("  Inputs: %d\n", len(tx.Inputs))
	fmt.Printf("    Input 0: %x:%d\n", tx.Inputs[0].PrevTxHash[:8], tx.Inputs[0].OutputIndex)
	fmt.Printf("  Outputs: %d\n", len(tx.Outputs))
	fmt.Printf("    Output 0: %.8f BTC to Bob\n", float64(tx.Outputs[0].Value)/100000000)
	fmt.Printf("    Output 1: %.8f BTC change to Alice\n", float64(tx.Outputs[1].Value)/100000000)
	fmt.Printf("  LockTime: %d\n", tx.LockTime)

	// Validate basic structure
	err = transaction.ValidateTransaction(tx)
	if err != nil {
		fmt.Printf("  ✗ Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Transaction structure is valid\n")
	}

	// Calculate transaction hash
	txHash, _ := serialization.HashTransaction(tx)
	fmt.Printf("  Transaction ID: %s\n", txHash)

	// Estimate size
	estimatedSize := transaction.CalculateSize(len(tx.Inputs), len(tx.Outputs))
	fmt.Printf("  Estimated size: %d bytes\n", estimatedSize)

	fmt.Println()
}

func demoCoinbaseTransaction() {
	fmt.Println("--- Demo 8: Coinbase Transaction ---")

	// Generate miner key
	minerKey, _ := keys.GeneratePrivateKey()
	minerAddress := minerKey.PublicKey().P2PKHAddress()

	fmt.Printf("Scenario: Miner mines a new block\n")
	fmt.Printf("Miner address: %s\n", minerAddress)

	// Create coinbase for block height 100
	blockHeight := uint64(100)
	blockReward := int64(5000000000) // 50 BTC (pre-halving)

	coinbase, err := transaction.CreateCoinbase(
		blockHeight,
		blockReward,
		minerAddress,
		[]byte("Mined with Learn-Bitcoin! Block #100"),
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("\nCoinbase transaction created:\n")
	fmt.Printf("  Block height: %d\n", blockHeight)
	fmt.Printf("  Block reward: %.8f BTC (%d satoshis)\n",
		float64(blockReward)/100000000, blockReward)
	fmt.Printf("  Is coinbase: %v\n", transaction.IsCoinbase(coinbase))

	// Show coinbase input
	fmt.Printf("\nCoinbase Input:\n")
	fmt.Printf("  Prev TX Hash: %x (all zeros)\n", coinbase.Inputs[0].PrevTxHash)
	fmt.Printf("  Output Index: 0x%x (0xFFFFFFFF)\n", coinbase.Inputs[0].OutputIndex)
	fmt.Printf("  ScriptSig: %x\n", coinbase.Inputs[0].SignatureScript)
	fmt.Printf("  Sequence: 0x%x\n", coinbase.Inputs[0].Sequence)

	// Show output
	fmt.Printf("\nCoinbase Output:\n")
	fmt.Printf("  Value: %.8f BTC\n", float64(coinbase.Outputs[0].Value)/100000000)
	fmt.Printf("  ScriptPubKey: %s\n", script.DisassembleScript(coinbase.Outputs[0].PubKeyScript))

	// Validate coinbase
	err = transaction.ValidateCoinbase(coinbase, blockHeight)
	if err != nil {
		fmt.Printf("  ✗ Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Coinbase is valid\n")
	}

	// Calculate transaction hash
	txHash, _ := serialization.HashTransaction(coinbase)
	fmt.Printf("  Transaction ID: %s\n", txHash)

	fmt.Println()
}

func demoFeeCalculation() {
	fmt.Println("--- Demo 9: Fee Calculation ---")

	// Create a transaction
	tx := &types.Transaction{
		Version: 1,
		Inputs: []types.TxInput{
			{PrevTxHash: types.Hash{1, 2, 3}, OutputIndex: 0},
			{PrevTxHash: types.Hash{4, 5, 6}, OutputIndex: 1},
		},
		Outputs: []types.TxOutput{
			{Value: 150000000}, // 1.5 BTC
			{Value: 50000000},  // 0.5 BTC
		},
		LockTime: 0,
	}

	// Previous outputs (what we're spending)
	prevOutputs := []types.TxOutput{
		{Value: 100000000}, // 1 BTC
		{Value: 110000000}, // 1.1 BTC
	}

	fmt.Printf("Transaction Analysis:\n")
	fmt.Printf("\nInputs (what we're spending):\n")
	fmt.Printf("  Input 1: %.8f BTC\n", float64(prevOutputs[0].Value)/100000000)
	fmt.Printf("  Input 2: %.8f BTC\n", float64(prevOutputs[1].Value)/100000000)
	totalIn := prevOutputs[0].Value + prevOutputs[1].Value
	fmt.Printf("  Total in: %.8f BTC (%d satoshis)\n", float64(totalIn)/100000000, totalIn)

	fmt.Printf("\nOutputs (where it's going):\n")
	fmt.Printf("  Output 1: %.8f BTC\n", float64(tx.Outputs[0].Value)/100000000)
	fmt.Printf("  Output 2: %.8f BTC\n", float64(tx.Outputs[1].Value)/100000000)
	totalOut := tx.Outputs[0].Value + tx.Outputs[1].Value
	fmt.Printf("  Total out: %.8f BTC (%d satoshis)\n", float64(totalOut)/100000000, totalOut)

	// Calculate fee
	fee, err := transaction.CalculateFee(tx, prevOutputs)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nFee Calculation:\n")
	fmt.Printf("  Fee: %.8f BTC (%d satoshis)\n", float64(fee)/100000000, fee)

	// Estimate size
	estimatedSize := transaction.CalculateSize(len(tx.Inputs), len(tx.Outputs))
	fmt.Printf("  Estimated size: %d bytes\n", estimatedSize)

	// Calculate fee rate
	feeRate := float64(fee) / float64(estimatedSize)
	fmt.Printf("  Fee rate: %.2f sat/byte\n", feeRate)

	// Show if fee is reasonable
	if feeRate < 1 {
		fmt.Printf("  ⚠ Warning: Fee rate very low, transaction may not confirm quickly\n")
	} else if feeRate < 10 {
		fmt.Printf("  ✓ Fee rate is reasonable\n")
	} else {
		fmt.Printf("  ⚠ Fee rate is high\n")
	}

	// Estimate fee for different fee rates
	fmt.Printf("\nFee estimates for different priorities:\n")
	fmt.Printf("  Low priority (1 sat/byte): %.8f BTC\n",
		float64(transaction.EstimateFee(2, 2, 1))/100000000)
	fmt.Printf("  Medium priority (10 sat/byte): %.8f BTC\n",
		float64(transaction.EstimateFee(2, 2, 10))/100000000)
	fmt.Printf("  High priority (50 sat/byte): %.8f BTC\n",
		float64(transaction.EstimateFee(2, 2, 50))/100000000)

	fmt.Println()
}

func demoSignatureHashTypes() {
	fmt.Println("--- Demo 10: Signature Hash Types ---")

	fmt.Println("Signature hash types determine what parts of a transaction are signed:")
	fmt.Println()

	// Show different hash types
	hashTypes := []transaction.SigHashType{
		transaction.SigHashAll,
		transaction.SigHashNone,
		transaction.SigHashSingle,
		transaction.SigHashAll | transaction.SigHashAnyOneCanPay,
		transaction.SigHashNone | transaction.SigHashAnyOneCanPay,
		transaction.SigHashSingle | transaction.SigHashAnyOneCanPay,
	}

	descriptions := []string{
		"Signs all inputs and all outputs (most common)",
		"Signs all inputs but no outputs (allows anyone to redirect payment)",
		"Signs all inputs and one output at same index",
		"Signs one input and all outputs (allows others to add inputs)",
		"Signs one input and no outputs",
		"Signs one input and one output",
	}

	for i, hashType := range hashTypes {
		fmt.Printf("%s (0x%02x):\n", transaction.SignatureHashInfo(hashType), hashType)
		fmt.Printf("  %s\n", descriptions[i])

		// Create a simple transaction for demonstration
		tx := &types.Transaction{
			Version: 1,
			Inputs: []types.TxInput{
				{PrevTxHash: types.Hash{1}, OutputIndex: 0},
				{PrevTxHash: types.Hash{2}, OutputIndex: 0},
			},
			Outputs: []types.TxOutput{
				{Value: 100000000},
				{Value: 50000000},
			},
			LockTime: 0,
		}

		// Calculate signature hash
		sigHash, err := transaction.CalcSignatureHash(tx, 0, []byte{0x76, 0xa9}, hashType)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Signature hash: %x...\n", sigHash[:8])
		}
		fmt.Println()
	}

	fmt.Println("Note: Different hash types enable different use cases:")
	fmt.Println("  - SIGHASH_ALL: Standard transactions")
	fmt.Println("  - SIGHASH_NONE: Blank checks")
	fmt.Println("  - SIGHASH_SINGLE: Assurance contracts")
	fmt.Println("  - ANYONECANPAY: Crowdfunding")

	fmt.Println()
}
