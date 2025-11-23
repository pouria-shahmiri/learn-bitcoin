package serialization

import (
	"bytes"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"io"
)

// SerializeTransaction converts transaction to bytes
// Order matters! Must match Bitcoin exactly
func SerializeTransaction(tx *types.Transaction) ([]byte, error) {
	var buf bytes.Buffer

	// 1. Version (4 bytes, little-endian)
	if err := WriteInt32(&buf, tx.Version); err != nil {
		return nil, err
	}

	// 2. Input count (VarInt)
	if err := WriteVarInt(&buf, uint64(len(tx.Inputs))); err != nil {
		return nil, err
	}

	// 3. Each input
	for _, input := range tx.Inputs {
		// Previous transaction hash (32 bytes)
		buf.Write(input.PrevTxHash[:])

		// Output index (4 bytes)
		if err := WriteUint32(&buf, input.OutputIndex); err != nil {
			return nil, err
		}

		// Signature script (VarInt length + data)
		if err := WriteBytes(&buf, input.SignatureScript); err != nil {
			return nil, err
		}

		// Sequence (4 bytes)
		if err := WriteUint32(&buf, input.Sequence); err != nil {
			return nil, err
		}
	}

	// 4. Output count (VarInt)
	if err := WriteVarInt(&buf, uint64(len(tx.Outputs))); err != nil {
		return nil, err
	}

	// 5. Each output
	for _, output := range tx.Outputs {
		// Value (8 bytes)
		if err := WriteUint64(&buf, uint64(output.Value)); err != nil {
			return nil, err
		}

		// Pubkey script (VarInt length + data)
		if err := WriteBytes(&buf, output.PubKeyScript); err != nil {
			return nil, err
		}
	}

	// 6. Locktime (4 bytes)
	if err := WriteUint32(&buf, tx.LockTime); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeTransaction reads transaction from bytes
func DeserializeTransaction(r io.Reader) (*types.Transaction, error) {
	var tx types.Transaction
	var err error

	if tx.Version, err = ReadInt32(r); err != nil {
		return nil, err
	}

	// Inputs
	inputCount, err := ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	tx.Inputs = make([]types.TxInput, inputCount)
	for i := uint64(0); i < inputCount; i++ {
		if _, err = io.ReadFull(r, tx.Inputs[i].PrevTxHash[:]); err != nil {
			return nil, err
		}
		if tx.Inputs[i].OutputIndex, err = ReadUint32(r); err != nil {
			return nil, err
		}
		if tx.Inputs[i].SignatureScript, err = ReadBytes(r); err != nil {
			return nil, err
		}
		if tx.Inputs[i].Sequence, err = ReadUint32(r); err != nil {
			return nil, err
		}
	}

	// Outputs
	outputCount, err := ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	tx.Outputs = make([]types.TxOutput, outputCount)
	for i := uint64(0); i < outputCount; i++ {
		val, err := ReadUint64(r)
		if err != nil {
			return nil, err
		}
		tx.Outputs[i].Value = int64(val)

		if tx.Outputs[i].PubKeyScript, err = ReadBytes(r); err != nil {
			return nil, err
		}
	}

	if tx.LockTime, err = ReadUint32(r); err != nil {
		return nil, err
	}

	return &tx, nil
}

// HashTransaction computes transaction ID
func HashTransaction(tx *types.Transaction) (types.Hash, error) {
	serialized, err := SerializeTransaction(tx)
	if err != nil {
		return types.Hash{}, err
	}
	return crypto.HashTransaction(serialized), nil
}

/*
```
**Transaction format (bytes):**
```
[Version: 4 bytes]
[Input count: VarInt]
  [Input 0]
    [Prev TX hash: 32 bytes]
    [Output index: 4 bytes]
    [Script length: VarInt]
    [Script: variable]
    [Sequence: 4 bytes]
  [Input 1...]
[Output count: VarInt]
  [Output 0]
    [Value: 8 bytes]
    [Script length: VarInt]
    [Script: variable]
  [Output 1...]
[Locktime: 4 bytes]
*/
