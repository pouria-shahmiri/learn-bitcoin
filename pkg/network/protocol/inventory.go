package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
)

// Inventory types
const (
	InvTypeError         = 0
	InvTypeTx            = 1 // Transaction
	InvTypeBlock         = 2 // Block
	InvTypeFilteredBlock = 3 // Filtered block (merkle block)
	InvTypeCompactBlock  = 4 // Compact block
)

// InvVect represents an inventory vector
type InvVect struct {
	Type uint32     // Inventory type
	Hash types.Hash // Object hash
}

// NewInvVect creates a new inventory vector
func NewInvVect(invType uint32, hash types.Hash) *InvVect {
	return &InvVect{
		Type: invType,
		Hash: hash,
	}
}

// InvMessage announces known objects
type InvMessage struct {
	Inventory []*InvVect
}

// NewInvMessage creates a new inv message
func NewInvMessage() *InvMessage {
	return &InvMessage{
		Inventory: make([]*InvVect, 0),
	}
}

// AddInvVect adds an inventory vector
func (inv *InvMessage) AddInvVect(vect *InvVect) {
	inv.Inventory = append(inv.Inventory, vect)
}

// Serialize converts inv message to bytes
func (inv *InvMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write count
	count := uint64(len(inv.Inventory))
	if err := writeVarInt(buf, count); err != nil {
		return nil, err
	}

	// Write inventory vectors
	for _, vect := range inv.Inventory {
		if err := binary.Write(buf, binary.LittleEndian, vect.Type); err != nil {
			return nil, err
		}
		if _, err := buf.Write(vect.Hash[:]); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// DeserializeInv reads an inv message from bytes
func DeserializeInv(data []byte) (*InvMessage, error) {
	buf := bytes.NewReader(data)
	inv := NewInvMessage()

	// Read count
	count, err := readVarInt(buf)
	if err != nil {
		return nil, err
	}

	// Read inventory vectors
	for i := uint64(0); i < count; i++ {
		vect := &InvVect{}

		if err := binary.Read(buf, binary.LittleEndian, &vect.Type); err != nil {
			return nil, err
		}

		if _, err := buf.Read(vect.Hash[:]); err != nil {
			return nil, err
		}

		inv.Inventory = append(inv.Inventory, vect)
	}

	return inv, nil
}

// GetDataMessage requests objects
type GetDataMessage struct {
	Inventory []*InvVect
}

// NewGetDataMessage creates a new getdata message
func NewGetDataMessage() *GetDataMessage {
	return &GetDataMessage{
		Inventory: make([]*InvVect, 0),
	}
}

// AddInvVect adds an inventory vector
func (gd *GetDataMessage) AddInvVect(vect *InvVect) {
	gd.Inventory = append(gd.Inventory, vect)
}

// Serialize converts getdata message to bytes
func (gd *GetDataMessage) Serialize() ([]byte, error) {
	// Same format as inv
	inv := &InvMessage{Inventory: gd.Inventory}
	return inv.Serialize()
}

// DeserializeGetData reads a getdata message from bytes
func DeserializeGetData(data []byte) (*GetDataMessage, error) {
	inv, err := DeserializeInv(data)
	if err != nil {
		return nil, err
	}
	return &GetDataMessage{Inventory: inv.Inventory}, nil
}

// GetBlocksMessage requests block hashes
type GetBlocksMessage struct {
	Version      uint32
	BlockLocator []types.Hash // Block hashes (newest to oldest)
	HashStop     types.Hash   // Stop at this hash (zero = get all)
}

// NewGetBlocksMessage creates a new getblocks message
func NewGetBlocksMessage(locator []types.Hash, hashStop types.Hash) *GetBlocksMessage {
	return &GetBlocksMessage{
		Version:      ProtocolVersion,
		BlockLocator: locator,
		HashStop:     hashStop,
	}
}

// Serialize converts getblocks message to bytes
func (gb *GetBlocksMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write version
	if err := binary.Write(buf, binary.LittleEndian, gb.Version); err != nil {
		return nil, err
	}

	// Write locator count
	count := uint64(len(gb.BlockLocator))
	if err := writeVarInt(buf, count); err != nil {
		return nil, err
	}

	// Write locator hashes
	for _, hash := range gb.BlockLocator {
		if _, err := buf.Write(hash[:]); err != nil {
			return nil, err
		}
	}

	// Write stop hash
	if _, err := buf.Write(gb.HashStop[:]); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeGetBlocks reads a getblocks message from bytes
func DeserializeGetBlocks(data []byte) (*GetBlocksMessage, error) {
	buf := bytes.NewReader(data)
	gb := &GetBlocksMessage{}

	// Read version
	if err := binary.Read(buf, binary.LittleEndian, &gb.Version); err != nil {
		return nil, err
	}

	// Read locator count
	count, err := readVarInt(buf)
	if err != nil {
		return nil, err
	}

	// Read locator hashes
	gb.BlockLocator = make([]types.Hash, count)
	for i := uint64(0); i < count; i++ {
		if _, err := buf.Read(gb.BlockLocator[i][:]); err != nil {
			return nil, err
		}
	}

	// Read stop hash
	if _, err := buf.Read(gb.HashStop[:]); err != nil {
		return nil, err
	}

	return gb, nil
}

// GetHeadersMessage requests block headers
type GetHeadersMessage struct {
	Version      uint32
	BlockLocator []types.Hash
	HashStop     types.Hash
}

// NewGetHeadersMessage creates a new getheaders message
func NewGetHeadersMessage(locator []types.Hash, hashStop types.Hash) *GetHeadersMessage {
	return &GetHeadersMessage{
		Version:      ProtocolVersion,
		BlockLocator: locator,
		HashStop:     hashStop,
	}
}

// Serialize converts getheaders message to bytes
func (gh *GetHeadersMessage) Serialize() ([]byte, error) {
	// Same format as getblocks
	gb := &GetBlocksMessage{
		Version:      gh.Version,
		BlockLocator: gh.BlockLocator,
		HashStop:     gh.HashStop,
	}
	return gb.Serialize()
}

// DeserializeGetHeaders reads a getheaders message from bytes
func DeserializeGetHeaders(data []byte) (*GetHeadersMessage, error) {
	gb, err := DeserializeGetBlocks(data)
	if err != nil {
		return nil, err
	}
	return &GetHeadersMessage{
		Version:      gb.Version,
		BlockLocator: gb.BlockLocator,
		HashStop:     gb.HashStop,
	}, nil
}

// String methods for debugging

func (inv *InvMessage) String() string {
	return fmt.Sprintf("Inv{Count: %d}", len(inv.Inventory))
}

func (gd *GetDataMessage) String() string {
	return fmt.Sprintf("GetData{Count: %d}", len(gd.Inventory))
}

func (gb *GetBlocksMessage) String() string {
	return fmt.Sprintf("GetBlocks{Locator: %d hashes}", len(gb.BlockLocator))
}

func (gh *GetHeadersMessage) String() string {
	return fmt.Sprintf("GetHeaders{Locator: %d hashes}", len(gh.BlockLocator))
}
