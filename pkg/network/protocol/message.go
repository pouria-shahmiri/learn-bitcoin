package protocol

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

// Bitcoin message format:
// [Magic Bytes (4)] [Command (12)] [Payload Length (4)] [Checksum (4)] [Payload (variable)]

const (
	// Magic bytes for mainnet
	MagicMainnet uint32 = 0xD9B4BEF9
	// Magic bytes for testnet
	MagicTestnet uint32 = 0x0709110B
	// Magic bytes for regtest (our testing network)
	MagicRegtest uint32 = 0xDAB5BFFA

	// Maximum payload size (32MB)
	MaxPayloadSize = 32 * 1024 * 1024

	// Command length
	CommandLength = 12
)

// Message types
const (
	CmdVersion    = "version"
	CmdVerAck     = "verack"
	CmdPing       = "ping"
	CmdPong       = "pong"
	CmdGetAddr    = "getaddr"
	CmdAddr       = "addr"
	CmdInv        = "inv"
	CmdGetData    = "getdata"
	CmdNotFound   = "notfound"
	CmdGetBlocks  = "getblocks"
	CmdGetHeaders = "getheaders"
	CmdTx         = "tx"
	CmdBlock      = "block"
	CmdHeaders    = "headers"
	CmdMempool    = "mempool"
	CmdReject     = "reject"
)

// Message represents a Bitcoin protocol message
type Message struct {
	Magic    uint32
	Command  string
	Payload  []byte
	Checksum [4]byte
}

// NewMessage creates a new message
func NewMessage(magic uint32, command string, payload []byte) *Message {
	msg := &Message{
		Magic:   magic,
		Command: command,
		Payload: payload,
	}
	msg.Checksum = msg.calculateChecksum()
	return msg
}

// calculateChecksum computes the checksum for the payload
func (m *Message) calculateChecksum() [4]byte {
	// Double SHA256 of payload
	first := sha256.Sum256(m.Payload)
	second := sha256.Sum256(first[:])

	var checksum [4]byte
	copy(checksum[:], second[:4])
	return checksum
}

// Serialize converts the message to bytes
func (m *Message) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write magic bytes
	if err := binary.Write(buf, binary.LittleEndian, m.Magic); err != nil {
		return nil, fmt.Errorf("failed to write magic: %w", err)
	}

	// Write command (12 bytes, null-padded)
	command := make([]byte, CommandLength)
	copy(command, m.Command)
	if _, err := buf.Write(command); err != nil {
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Write payload length
	payloadLen := uint32(len(m.Payload))
	if err := binary.Write(buf, binary.LittleEndian, payloadLen); err != nil {
		return nil, fmt.Errorf("failed to write payload length: %w", err)
	}

	// Write checksum
	if _, err := buf.Write(m.Checksum[:]); err != nil {
		return nil, fmt.Errorf("failed to write checksum: %w", err)
	}

	// Write payload
	if _, err := buf.Write(m.Payload); err != nil {
		return nil, fmt.Errorf("failed to write payload: %w", err)
	}

	return buf.Bytes(), nil
}

// Deserialize reads a message from bytes
func Deserialize(r io.Reader) (*Message, error) {
	msg := &Message{}

	// Read magic bytes
	if err := binary.Read(r, binary.LittleEndian, &msg.Magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}

	// Read command
	command := make([]byte, CommandLength)
	if _, err := io.ReadFull(r, command); err != nil {
		return nil, fmt.Errorf("failed to read command: %w", err)
	}
	// Remove null padding
	msg.Command = string(bytes.TrimRight(command, "\x00"))

	// Read payload length
	var payloadLen uint32
	if err := binary.Read(r, binary.LittleEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}

	// Validate payload length
	if payloadLen > MaxPayloadSize {
		return nil, fmt.Errorf("payload too large: %d bytes", payloadLen)
	}

	// Read checksum
	if _, err := io.ReadFull(r, msg.Checksum[:]); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	// Read payload
	if payloadLen > 0 {
		msg.Payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(r, msg.Payload); err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	}

	// Verify checksum
	expectedChecksum := msg.calculateChecksum()
	if msg.Checksum != expectedChecksum {
		return nil, fmt.Errorf("checksum mismatch: got %x, expected %x", msg.Checksum, expectedChecksum)
	}

	return msg, nil
}

// String returns a string representation of the message
func (m *Message) String() string {
	return fmt.Sprintf("Message{Command: %s, PayloadSize: %d}", m.Command, len(m.Payload))
}
