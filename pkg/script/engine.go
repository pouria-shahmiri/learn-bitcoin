package script

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/pouria-shahmiri/learn-bitcoin/pkg/types"
	"golang.org/x/crypto/ripemd160"
)

// Engine executes Bitcoin scripts
type Engine struct {
	stack    *Stack
	altStack *Stack
	script   []byte
	pc       int         // Program counter
	tx       interface{} // Transaction being validated
	inputIdx int         // Input index being validated
}

// NewEngine creates a new script execution engine
func NewEngine(script []byte) *Engine {
	return &Engine{
		stack:    NewStack(),
		altStack: NewStack(),
		script:   script,
		pc:       0,
	}
}

// Execute runs the script
func (e *Engine) Execute() error {
	for e.pc < len(e.script) {
		if err := e.step(); err != nil {
			return fmt.Errorf("execution failed at pc=%d: %w", e.pc, err)
		}
	}

	// Script succeeds if stack top is true
	if e.stack.Size() == 0 {
		return fmt.Errorf("script failed: empty stack")
	}

	top, err := e.stack.Peek()
	if err != nil {
		return err
	}

	if !castToBool(top) {
		return fmt.Errorf("script failed: false on stack")
	}

	return nil
}

// step executes one opcode
func (e *Engine) step() error {
	if e.pc >= len(e.script) {
		return fmt.Errorf("program counter out of bounds")
	}

	opcode := e.script[e.pc]
	e.pc++

	// Handle data push opcodes (0x01-0x4b push that many bytes)
	if opcode > 0 && opcode <= 0x4b {
		return e.executePush(int(opcode))
	}

	// Handle specific opcodes
	switch opcode {
	case OP_0:
		e.stack.Push([]byte{})

	case OP_1NEGATE:
		e.stack.PushInt(-1)

	case OP_1, OP_2, OP_3, OP_4, OP_5, OP_6, OP_7, OP_8,
		OP_9, OP_10, OP_11, OP_12, OP_13, OP_14, OP_15, OP_16:
		e.stack.PushInt(int64(SmallIntValue(opcode)))

	case OP_NOP:
		// Do nothing

	case OP_VERIFY:
		return e.opVerify()

	case OP_RETURN:
		return fmt.Errorf("OP_RETURN executed")

	case OP_DUP:
		return e.stack.Dup()

	case OP_EQUAL:
		return e.opEqual()

	case OP_EQUALVERIFY:
		if err := e.opEqual(); err != nil {
			return err
		}
		return e.opVerify()

	case OP_HASH160:
		return e.opHash160()

	case OP_SHA256:
		return e.opSHA256()

	case OP_CHECKSIG:
		return e.opCheckSig()

	case OP_CHECKSIGVERIFY:
		if err := e.opCheckSig(); err != nil {
			return err
		}
		return e.opVerify()

	case OP_DROP:
		_, err := e.stack.Pop()
		return err

	case OP_SWAP:
		return e.stack.Swap()

	default:
		return fmt.Errorf("unimplemented opcode: %s", OpcodeName(opcode))
	}

	return nil
}

// executePush pushes N bytes onto stack
func (e *Engine) executePush(n int) error {
	if e.pc+n > len(e.script) {
		return fmt.Errorf("push %d bytes exceeds script length", n)
	}

	data := make([]byte, n)
	copy(data, e.script[e.pc:e.pc+n])
	e.pc += n

	e.stack.Push(data)
	return nil
}

// opVerify pops top and fails if false
func (e *Engine) opVerify() error {
	item, err := e.stack.Pop()
	if err != nil {
		return err
	}

	if !castToBool(item) {
		return fmt.Errorf("VERIFY failed")
	}

	return nil
}

// opEqual pops two items and pushes true if equal
func (e *Engine) opEqual() error {
	a, err := e.stack.Pop()
	if err != nil {
		return err
	}

	b, err := e.stack.Pop()
	if err != nil {
		return err
	}

	if bytes.Equal(a, b) {
		e.stack.Push([]byte{1})
	} else {
		e.stack.Push([]byte{})
	}

	return nil
}

// opHash160 performs RIPEMD160(SHA256(x))
func (e *Engine) opHash160() error {
	item, err := e.stack.Pop()
	if err != nil {
		return err
	}

	// SHA256
	sha := sha256.Sum256(item)

	// RIPEMD160
	ripe := ripemd160.New()
	ripe.Write(sha[:])
	hash := ripe.Sum(nil)

	e.stack.Push(hash)
	return nil
}

// opSHA256 performs SHA256(x)
func (e *Engine) opSHA256() error {
	item, err := e.stack.Pop()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(item)
	e.stack.Push(hash[:])

	return nil
}

// opCheckSig verifies signature
func (e *Engine) opCheckSig() error {
	// Pop public key
	pubKeyBytes, err := e.stack.Pop()
	if err != nil {
		return err
	}

	// Pop signature
	sigBytes, err := e.stack.Pop()
	if err != nil {
		return err
	}

	// Basic validation - ensure we have data
	if len(pubKeyBytes) == 0 || len(sigBytes) == 0 {
		e.stack.Push([]byte{})
		return nil
	}

	// For now, push true (we'll implement full validation later)
	// Full implementation requires transaction context
	e.stack.Push([]byte{1})

	return nil
}

// castToBool converts script item to boolean
func castToBool(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != 0 {
			// Check for negative zero
			if i == len(b)-1 && b[i] == 0x80 {
				return false
			}
			return true
		}
	}
	return false
}

// Stack returns the main stack (for debugging)
func (e *Engine) Stack() *Stack {
	return e.stack
}

// SetTransaction sets the transaction context for signature verification
func (e *Engine) SetTransaction(tx *types.Transaction, inputIdx int) {
	e.tx = tx
	e.inputIdx = inputIdx
}
