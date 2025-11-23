package script

import (
	"fmt"
)

// Stack represents a LIFO stack for script execution
type Stack struct {
	data [][]byte
}

// NewStack creates a new empty stack
func NewStack() *Stack {
	return &Stack{
		data: make([][]byte, 0),
	}
}

// Push adds an item to the top of the stack
func (s *Stack) Push(item []byte) {
	s.data = append(s.data, item)
}

// Pop removes and returns the top item
func (s *Stack) Pop() ([]byte, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("stack underflow")
	}

	item := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	return item, nil
}

// Peek returns the top item without removing it
func (s *Stack) Peek() ([]byte, error) {
	if len(s.data) == 0 {
		return nil, fmt.Errorf("stack empty")
	}

	return s.data[len(s.data)-1], nil
}

// PeekN returns the Nth item from the top (0 = top)
func (s *Stack) PeekN(n int) ([]byte, error) {
	if n < 0 || n >= len(s.data) {
		return nil, fmt.Errorf("stack index out of range: %d", n)
	}

	return s.data[len(s.data)-1-n], nil
}

// Dup duplicates the top stack item
func (s *Stack) Dup() error {
	item, err := s.Peek()
	if err != nil {
		return err
	}

	// Make a copy
	dup := make([]byte, len(item))
	copy(dup, item)

	s.Push(dup)
	return nil
}

// Swap swaps the top two stack items
func (s *Stack) Swap() error {
	if len(s.data) < 2 {
		return fmt.Errorf("stack has fewer than 2 items")
	}

	top := len(s.data) - 1
	s.data[top], s.data[top-1] = s.data[top-1], s.data[top]

	return nil
}

// Size returns the number of items on the stack
func (s *Stack) Size() int {
	return len(s.data)
}

// Clear removes all items from the stack
func (s *Stack) Clear() {
	s.data = s.data[:0]
}

// String returns a string representation of the stack
func (s *Stack) String() string {
	if len(s.data) == 0 {
		return "[]"
	}

	result := "["
	for i := len(s.data) - 1; i >= 0; i-- {
		result += fmt.Sprintf("%x", s.data[i])
		if i > 0 {
			result += ", "
		}
	}
	result += "]"

	return result
}

// AsInt converts top stack item to int64
func (s *Stack) AsInt() (int64, error) {
	item, err := s.Peek()
	if err != nil {
		return 0, err
	}

	return scriptNumToInt64(item), nil
}

// PushInt pushes an integer as script number
func (s *Stack) PushInt(n int64) {
	s.Push(int64ToScriptNum(n))
}

// Helper: Convert script number to int64
func scriptNumToInt64(b []byte) int64 {
	if len(b) == 0 {
		return 0
	}

	// Check if negative (high bit of last byte)
	negative := (b[len(b)-1] & 0x80) != 0

	// Convert bytes to int64 (little-endian)
	result := int64(0)
	for i := 0; i < len(b); i++ {
		val := int64(b[i])
		if i == len(b)-1 && negative {
			val &= 0x7f // Clear sign bit
		}
		result |= val << uint(8*i)
	}

	if negative {
		result = -result
	}

	return result
}

// Helper: Convert int64 to script number
func int64ToScriptNum(n int64) []byte {
	if n == 0 {
		return []byte{}
	}

	negative := n < 0
	if negative {
		n = -n
	}

	// Convert to bytes (little-endian)
	result := []byte{}
	for n > 0 {
		result = append(result, byte(n&0xff))
		n >>= 8
	}

	// Add sign bit if needed
	if (result[len(result)-1] & 0x80) != 0 {
		// High bit is set, need extra byte
		if negative {
			result = append(result, 0x80)
		} else {
			result = append(result, 0x00)
		}
	} else if negative {
		// Set high bit for negative
		result[len(result)-1] |= 0x80
	}

	return result
}
