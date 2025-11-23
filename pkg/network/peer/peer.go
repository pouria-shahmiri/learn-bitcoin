package peer

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/network/protocol"
)

// Peer represents a connected node
type Peer struct {
	Conn           net.Conn
	addr           string
	Inbound        bool // True if peer connected to us, false if we connected to them
	ConnectedAt    time.Time
	LastActive     time.Time
	VerAckReceived bool
	Version        *protocol.VersionMessage

	// Channels for communication
	Send    chan *protocol.Message
	Receive chan *protocol.Message
	Quit    chan struct{}

	wg sync.WaitGroup
}

// NewPeer creates a new peer instance
func NewPeer(conn net.Conn, inbound bool) *Peer {
	return &Peer{
		Conn:        conn,
		addr:        conn.RemoteAddr().String(),
		Inbound:     inbound,
		ConnectedAt: time.Now(),
		LastActive:  time.Now(),
		Send:        make(chan *protocol.Message, 100),
		Receive:     make(chan *protocol.Message, 100),
		Quit:        make(chan struct{}),
	}
}

// Start begins the read/write loops
func (p *Peer) Start() {
	p.wg.Add(2)
	go p.readLoop()
	go p.writeLoop()
}

// Stop terminates the connection
func (p *Peer) Stop() {
	close(p.Quit)
	p.Conn.Close()
	p.wg.Wait()
}

// SendMessage queues a message to be sent
func (p *Peer) SendMessage(msg *protocol.Message) {
	select {
	case p.Send <- msg:
	case <-p.Quit:
	}
}

// readLoop reads messages from the connection
func (p *Peer) readLoop() {
	defer p.wg.Done()

	reader := bufio.NewReader(p.Conn)

	for {
		select {
		case <-p.Quit:
			return
		default:
			// Set read deadline
			p.Conn.SetReadDeadline(time.Now().Add(20 * time.Minute))

			// Read message
			msg, err := protocol.Deserialize(reader)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Error reading from peer %s: %v\n", p.addr, err)
				}
				// Close connection on error
				// In a real app we'd signal the manager to remove this peer
				return
			}

			p.LastActive = time.Now()

			// Send to receive channel
			select {
			case p.Receive <- msg:
			case <-p.Quit:
				return
			}
		}
	}
}

// writeLoop writes messages to the connection
func (p *Peer) writeLoop() {
	defer p.wg.Done()

	for {
		select {
		case msg := <-p.Send:
			// Set write deadline
			p.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

			serialized, err := msg.Serialize()
			if err != nil {
				fmt.Printf("Error serializing message: %v\n", err)
				continue
			}

			if _, err := p.Conn.Write(serialized); err != nil {
				fmt.Printf("Error writing to peer %s: %v\n", p.addr, err)
				return
			}

		case <-p.Quit:
			return
		}
	}
}

// Handshake performs the version negotiation
func (p *Peer) Handshake(localVersion *protocol.VersionMessage) error {
	// 1. Send Version
	msg := protocol.NewMessage(protocol.MagicMainnet, protocol.CmdVersion, mustSerialize(localVersion))
	p.SendMessage(msg)

	// 2. Wait for Version (handled in read loop, but for simplicity in this phase we might want to block or handle async)
	// In a full implementation, this is state machine driven.
	// For this phase, we'll assume the upper layer handles the response logic.

	return nil
}

func mustSerialize(v *protocol.VersionMessage) []byte {
	b, err := v.Serialize()
	if err != nil {
		panic(err)
	}
	return b
}

// Address returns the peer's address
func (p *Peer) Address() string {
	return p.addr
}
