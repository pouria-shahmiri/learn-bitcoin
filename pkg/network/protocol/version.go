package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// Protocol version
const (
	ProtocolVersion    = 70015 // Bitcoin Core 0.13.2+
	MinProtocolVersion = 70001
)

// Service flags
const (
	SFNodeNetwork        = 1 << 0  // Full node (can serve full blocks)
	SFNodeGetUTXO        = 1 << 1  // Supports getutxo
	SFNodeBloom          = 1 << 2  // Supports bloom filtering
	SFNodeWitness        = 1 << 3  // Supports segregated witness
	SFNodeNetworkLimited = 1 << 10 // Pruned node with limited history
)

// NetAddress represents a network address
type NetAddress struct {
	Services uint64   // Service flags
	IP       [16]byte // IPv6 address (IPv4 mapped)
	Port     uint16   // Port number
}

// VersionMessage is sent during handshake
type VersionMessage struct {
	Version     int32      // Protocol version
	Services    uint64     // Service flags
	Timestamp   int64      // Current timestamp
	AddrRecv    NetAddress // Address of receiving node
	AddrFrom    NetAddress // Address of sending node
	Nonce       uint64     // Random nonce (to detect self-connections)
	UserAgent   string     // Client identification
	StartHeight int32      // Last block height known
	Relay       bool       // Whether to relay transactions
}

// NewVersionMessage creates a new version message
func NewVersionMessage(addrRecv, addrFrom NetAddress, nonce uint64, userAgent string, startHeight int32) *VersionMessage {
	return &VersionMessage{
		Version:     ProtocolVersion,
		Services:    SFNodeNetwork,
		Timestamp:   time.Now().Unix(),
		AddrRecv:    addrRecv,
		AddrFrom:    addrFrom,
		Nonce:       nonce,
		UserAgent:   userAgent,
		StartHeight: startHeight,
		Relay:       true,
	}
}

// Serialize converts version message to bytes
func (v *VersionMessage) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write version
	if err := binary.Write(buf, binary.LittleEndian, v.Version); err != nil {
		return nil, err
	}

	// Write services
	if err := binary.Write(buf, binary.LittleEndian, v.Services); err != nil {
		return nil, err
	}

	// Write timestamp
	if err := binary.Write(buf, binary.LittleEndian, v.Timestamp); err != nil {
		return nil, err
	}

	// Write receiver address
	if err := writeNetAddress(buf, v.AddrRecv); err != nil {
		return nil, err
	}

	// Write sender address
	if err := writeNetAddress(buf, v.AddrFrom); err != nil {
		return nil, err
	}

	// Write nonce
	if err := binary.Write(buf, binary.LittleEndian, v.Nonce); err != nil {
		return nil, err
	}

	// Write user agent (variable length string)
	if err := writeVarString(buf, v.UserAgent); err != nil {
		return nil, err
	}

	// Write start height
	if err := binary.Write(buf, binary.LittleEndian, v.StartHeight); err != nil {
		return nil, err
	}

	// Write relay flag
	relay := byte(0)
	if v.Relay {
		relay = 1
	}
	if err := buf.WriteByte(relay); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DeserializeVersion reads a version message from bytes
func DeserializeVersion(data []byte) (*VersionMessage, error) {
	buf := bytes.NewReader(data)
	v := &VersionMessage{}

	// Read version
	if err := binary.Read(buf, binary.LittleEndian, &v.Version); err != nil {
		return nil, err
	}

	// Read services
	if err := binary.Read(buf, binary.LittleEndian, &v.Services); err != nil {
		return nil, err
	}

	// Read timestamp
	if err := binary.Read(buf, binary.LittleEndian, &v.Timestamp); err != nil {
		return nil, err
	}

	// Read receiver address
	addrRecv, err := readNetAddress(buf)
	if err != nil {
		return nil, err
	}
	v.AddrRecv = addrRecv

	// Read sender address
	addrFrom, err := readNetAddress(buf)
	if err != nil {
		return nil, err
	}
	v.AddrFrom = addrFrom

	// Read nonce
	if err := binary.Read(buf, binary.LittleEndian, &v.Nonce); err != nil {
		return nil, err
	}

	// Read user agent
	userAgent, err := readVarString(buf)
	if err != nil {
		return nil, err
	}
	v.UserAgent = userAgent

	// Read start height
	if err := binary.Read(buf, binary.LittleEndian, &v.StartHeight); err != nil {
		return nil, err
	}

	// Read relay flag (optional, default true)
	relay := make([]byte, 1)
	if _, err := buf.Read(relay); err == nil {
		v.Relay = relay[0] != 0
	} else {
		v.Relay = true
	}

	return v, nil
}

// Helper functions

func writeNetAddress(buf *bytes.Buffer, addr NetAddress) error {
	if err := binary.Write(buf, binary.LittleEndian, addr.Services); err != nil {
		return err
	}
	if _, err := buf.Write(addr.IP[:]); err != nil {
		return err
	}
	// Port is big-endian
	if err := binary.Write(buf, binary.BigEndian, addr.Port); err != nil {
		return err
	}
	return nil
}

func readNetAddress(buf *bytes.Reader) (NetAddress, error) {
	var addr NetAddress

	if err := binary.Read(buf, binary.LittleEndian, &addr.Services); err != nil {
		return addr, err
	}
	if _, err := buf.Read(addr.IP[:]); err != nil {
		return addr, err
	}
	// Port is big-endian
	if err := binary.Read(buf, binary.BigEndian, &addr.Port); err != nil {
		return addr, err
	}

	return addr, nil
}

func writeVarString(buf *bytes.Buffer, s string) error {
	// Write length as varint
	if err := writeVarInt(buf, uint64(len(s))); err != nil {
		return err
	}
	// Write string
	if _, err := buf.WriteString(s); err != nil {
		return err
	}
	return nil
}

func readVarString(buf *bytes.Reader) (string, error) {
	// Read length
	length, err := readVarInt(buf)
	if err != nil {
		return "", err
	}

	// Read string
	str := make([]byte, length)
	if _, err := buf.Read(str); err != nil {
		return "", err
	}

	return string(str), nil
}

func writeVarInt(buf *bytes.Buffer, n uint64) error {
	if n < 0xFD {
		return buf.WriteByte(byte(n))
	} else if n <= 0xFFFF {
		if err := buf.WriteByte(0xFD); err != nil {
			return err
		}
		return binary.Write(buf, binary.LittleEndian, uint16(n))
	} else if n <= 0xFFFFFFFF {
		if err := buf.WriteByte(0xFE); err != nil {
			return err
		}
		return binary.Write(buf, binary.LittleEndian, uint32(n))
	} else {
		if err := buf.WriteByte(0xFF); err != nil {
			return err
		}
		return binary.Write(buf, binary.LittleEndian, n)
	}
}

func readVarInt(buf *bytes.Reader) (uint64, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}

	switch b {
	case 0xFF:
		var n uint64
		err := binary.Read(buf, binary.LittleEndian, &n)
		return n, err
	case 0xFE:
		var n uint32
		err := binary.Read(buf, binary.LittleEndian, &n)
		return uint64(n), err
	case 0xFD:
		var n uint16
		err := binary.Read(buf, binary.LittleEndian, &n)
		return uint64(n), err
	default:
		return uint64(b), nil
	}
}

// String returns a string representation
func (v *VersionMessage) String() string {
	return fmt.Sprintf("Version{Version: %d, Services: %d, UserAgent: %s, Height: %d}",
		v.Version, v.Services, v.UserAgent, v.StartHeight)
}
