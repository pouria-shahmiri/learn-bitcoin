package utxo

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash      types.Hash     // Transaction hash
	OutputIndex uint32         // Index of output in transaction
	Output      types.TxOutput // The actual output
	Height      uint64         // Block height where it was created
	IsCoinbase  bool           // Is this from a coinbase transaction?
}

// OutPoint uniquely identifies a transaction output
type OutPoint struct {
	Hash  types.Hash
	Index uint32
}

// NewOutPoint creates a new outpoint
func NewOutPoint(hash types.Hash, index uint32) OutPoint {
	return OutPoint{
		Hash:  hash,
		Index: index,
	}
}

// String returns string representation of outpoint
func (op OutPoint) String() string {
	return fmt.Sprintf("%s:%d", op.Hash, op.Index)
}

// Bytes serializes outpoint to bytes
func (op OutPoint) Bytes() []byte {
	buf := make([]byte, 36) // 32 bytes hash + 4 bytes index
	copy(buf[0:32], op.Hash[:])
	binary.LittleEndian.PutUint32(buf[32:36], op.Index)
	return buf
}

// OutPointFromBytes deserializes outpoint from bytes
func OutPointFromBytes(data []byte) (OutPoint, error) {
	if len(data) != 36 {
		return OutPoint{}, fmt.Errorf("invalid outpoint data length: %d", len(data))
	}

	var op OutPoint
	copy(op.Hash[:], data[0:32])
	op.Index = binary.LittleEndian.Uint32(data[32:36])

	return op, nil
}

// Equal checks if two outpoints are equal
func (op OutPoint) Equal(other OutPoint) bool {
	return op.Hash == other.Hash && op.Index == other.Index
}

// IsNull checks if outpoint is null (all zeros)
func (op OutPoint) IsNull() bool {
	for _, b := range op.Hash {
		if b != 0 {
			return false
		}
	}
	return op.Index == 0xFFFFFFFF
}

// NewUTXO creates a new UTXO
func NewUTXO(txHash types.Hash, outputIndex uint32, output types.TxOutput, height uint64, isCoinbase bool) *UTXO {
	return &UTXO{
		TxHash:      txHash,
		OutputIndex: outputIndex,
		Output:      output,
		Height:      height,
		IsCoinbase:  isCoinbase,
	}
}

// OutPoint returns the outpoint for this UTXO
func (u *UTXO) OutPoint() OutPoint {
	return NewOutPoint(u.TxHash, u.OutputIndex)
}

// Value returns the value of this UTXO
func (u *UTXO) Value() int64 {
	return u.Output.Value
}

// IsMature checks if coinbase UTXO is mature (can be spent)
// Bitcoin requires 100 confirmations for coinbase outputs
func (u *UTXO) IsMature(currentHeight uint64) bool {
	if !u.IsCoinbase {
		return true
	}

	const coinbaseMaturity = 100
	return currentHeight >= u.Height+coinbaseMaturity
}

// Serialize converts UTXO to bytes for storage
func (u *UTXO) Serialize() []byte {
	var buf bytes.Buffer

	// Transaction hash (32 bytes)
	buf.Write(u.TxHash[:])

	// Output index (4 bytes)
	binary.Write(&buf, binary.LittleEndian, u.OutputIndex)

	// Value (8 bytes)
	binary.Write(&buf, binary.LittleEndian, u.Output.Value)

	// Script length (varint)
	scriptLen := uint64(len(u.Output.PubKeyScript))
	binary.Write(&buf, binary.LittleEndian, scriptLen)

	// Script
	buf.Write(u.Output.PubKeyScript)

	// Height (8 bytes)
	binary.Write(&buf, binary.LittleEndian, u.Height)

	// Is coinbase (1 byte)
	if u.IsCoinbase {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}

	return buf.Bytes()
}

// DeserializeUTXO converts bytes back to UTXO
func DeserializeUTXO(data []byte) (*UTXO, error) {
	if len(data) < 53 { // Minimum size: 32+4+8+8+1
		return nil, fmt.Errorf("data too short for UTXO")
	}

	buf := bytes.NewReader(data)
	utxo := &UTXO{}

	// Transaction hash
	buf.Read(utxo.TxHash[:])

	// Output index
	binary.Read(buf, binary.LittleEndian, &utxo.OutputIndex)

	// Value
	binary.Read(buf, binary.LittleEndian, &utxo.Output.Value)

	// Script length
	var scriptLen uint64
	binary.Read(buf, binary.LittleEndian, &scriptLen)

	// Script
	utxo.Output.PubKeyScript = make([]byte, scriptLen)
	buf.Read(utxo.Output.PubKeyScript)

	// Height
	binary.Read(buf, binary.LittleEndian, &utxo.Height)

	// Is coinbase
	coinbaseByte, _ := buf.ReadByte()
	utxo.IsCoinbase = coinbaseByte == 1

	return utxo, nil
}

// Clone creates a deep copy of UTXO
func (u *UTXO) Clone() *UTXO {
	clone := &UTXO{
		TxHash:      u.TxHash,
		OutputIndex: u.OutputIndex,
		Output: types.TxOutput{
			Value: u.Output.Value,
		},
		Height:     u.Height,
		IsCoinbase: u.IsCoinbase,
	}

	// Deep copy script
	clone.Output.PubKeyScript = make([]byte, len(u.Output.PubKeyScript))
	copy(clone.Output.PubKeyScript, u.Output.PubKeyScript)

	return clone
}
