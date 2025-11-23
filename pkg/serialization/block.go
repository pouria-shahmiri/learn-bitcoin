package serialization

import (
	"bytes"
	"io"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/crypto"
	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// SerializeBlockHeader converts header to 80 bytes
func SerializeBlockHeader(bh *types.BlockHeader) ([]byte, error) {
	var buf bytes.Buffer

	// Version (4 bytes)
	if err := WriteInt32(&buf, bh.Version); err != nil {
		return nil, err
	}

	// Previous block hash (32 bytes)
	buf.Write(bh.PrevBlockHash[:])

	// Merkle root (32 bytes)
	buf.Write(bh.MerkleRoot[:])

	// Timestamp (4 bytes)
	if err := WriteUint32(&buf, bh.Timestamp); err != nil {
		return nil, err
	}

	// Difficulty bits (4 bytes)
	if err := WriteUint32(&buf, bh.Bits); err != nil {
		return nil, err
	}

	// Nonce (4 bytes)
	if err := WriteUint32(&buf, bh.Nonce); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeBlockHeader reads 80 bytes into header
func DeserializeBlockHeader(r io.Reader) (*types.BlockHeader, error) {
	var bh types.BlockHeader
	var err error

	if bh.Version, err = ReadInt32(r); err != nil {
		return nil, err
	}

	if _, err = io.ReadFull(r, bh.PrevBlockHash[:]); err != nil {
		return nil, err
	}
	if _, err = io.ReadFull(r, bh.MerkleRoot[:]); err != nil {
		return nil, err
	}

	if bh.Timestamp, err = ReadUint32(r); err != nil {
		return nil, err
	}
	if bh.Bits, err = ReadUint32(r); err != nil {
		return nil, err
	}
	if bh.Nonce, err = ReadUint32(r); err != nil {
		return nil, err
	}

	return &bh, nil
}

// HashBlockHeader computes block hash
func HashBlockHeader(bh *types.BlockHeader) (types.Hash, error) {
	serialized, err := SerializeBlockHeader(bh)
	if err != nil {
		return types.Hash{}, err
	}

	// Sanity check
	if len(serialized) != 80 {
		panic("block header must be exactly 80 bytes")
	}

	return crypto.HashBlockHeader(serialized), nil
}

// SerializeBlock serializes a complete block
func SerializeBlock(block *types.Block) ([]byte, error) {
	var buf bytes.Buffer

	// Serialize header
	headerBytes, err := SerializeBlockHeader(&block.Header)
	if err != nil {
		return nil, err
	}
	buf.Write(headerBytes)

	// Transaction count
	if err := WriteVarInt(&buf, uint64(len(block.Transactions))); err != nil {
		return nil, err
	}

	// Serialize each transaction
	for _, tx := range block.Transactions {
		txBytes, err := SerializeTransaction(&tx)
		if err != nil {
			return nil, err
		}
		buf.Write(txBytes)
	}

	return buf.Bytes(), nil
}

// DeserializeBlock reads a complete block
func DeserializeBlock(data []byte) (*types.Block, error) {
	r := bytes.NewReader(data)

	header, err := DeserializeBlockHeader(r)
	if err != nil {
		return nil, err
	}

	txCount, err := ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	txs := make([]types.Transaction, txCount)
	for i := uint64(0); i < txCount; i++ {
		tx, err := DeserializeTransaction(r)
		if err != nil {
			return nil, err
		}
		txs[i] = *tx
	}

	return &types.Block{
		Header:       *header,
		Transactions: txs,
	}, nil
}
