package script

import (
	"encoding/binary"
)

// Builder helps construct scripts
type Builder struct {
	script []byte
}

// NewBuilder creates a new script builder
func NewBuilder() *Builder {
	return &Builder{
		script: []byte{},
	}
}

// AddOp adds an opcode
func (b *Builder) AddOp(opcode byte) *Builder {
	b.script = append(b.script, opcode)
	return b
}

// AddData pushes data onto the stack
func (b *Builder) AddData(data []byte) *Builder {
	length := len(data)

	if length == 0 {
		b.script = append(b.script, OP_0)
		return b
	}

	if length <= 75 {
		// Direct push (0x01-0x4b)
		b.script = append(b.script, byte(length))
		b.script = append(b.script, data...)
	} else if length <= 0xff {
		// OP_PUSHDATA1
		b.script = append(b.script, OP_PUSHDATA1, byte(length))
		b.script = append(b.script, data...)
	} else if length <= 0xffff {
		// OP_PUSHDATA2
		b.script = append(b.script, OP_PUSHDATA2)
		lenBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBytes, uint16(length))
		b.script = append(b.script, lenBytes...)
		b.script = append(b.script, data...)
	} else {
		// OP_PUSHDATA4
		b.script = append(b.script, OP_PUSHDATA4)
		lenBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenBytes, uint32(length))
		b.script = append(b.script, lenBytes...)
		b.script = append(b.script, data...)
	}

	return b
}

// AddInt pushes an integer
func (b *Builder) AddInt(n int64) *Builder {
	if n == -1 {
		b.script = append(b.script, OP_1NEGATE)
		return b
	}

	if n == 0 {
		b.script = append(b.script, OP_0)
		return b
	}

	if n >= 1 && n <= 16 {
		b.script = append(b.script, byte(OP_1+n-1))
		return b
	}

	// Use script number encoding
	return b.AddData(int64ToScriptNum(n))
}

// Script returns the built script
func (b *Builder) Script() []byte {
	return b.script
}

// Reset clears the builder
func (b *Builder) Reset() *Builder {
	b.script = b.script[:0]
	return b
}
