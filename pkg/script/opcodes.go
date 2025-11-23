package script

import "fmt"

// Bitcoin Script opcodes
const (
	// Constants
	OP_0         = 0x00 // Also known as OP_FALSE
	OP_FALSE     = 0x00
	OP_PUSHDATA1 = 0x4c // Push next byte as N, then push N bytes
	OP_PUSHDATA2 = 0x4d // Push next 2 bytes as N, then push N bytes
	OP_PUSHDATA4 = 0x4e // Push next 4 bytes as N, then push N bytes
	OP_1NEGATE   = 0x4f // Push -1
	OP_1         = 0x51 // Push 1
	OP_TRUE      = 0x51
	OP_2         = 0x52
	OP_3         = 0x53
	OP_4         = 0x54
	OP_5         = 0x55
	OP_6         = 0x56
	OP_7         = 0x57
	OP_8         = 0x58
	OP_9         = 0x59
	OP_10        = 0x5a
	OP_11        = 0x5b
	OP_12        = 0x5c
	OP_13        = 0x5d
	OP_14        = 0x5e
	OP_15        = 0x5f
	OP_16        = 0x60

	// Flow control
	OP_NOP    = 0x61
	OP_IF     = 0x63
	OP_NOTIF  = 0x64
	OP_ELSE   = 0x67
	OP_ENDIF  = 0x68
	OP_VERIFY = 0x69
	OP_RETURN = 0x6a

	// Stack
	OP_TOALTSTACK   = 0x6b
	OP_FROMALTSTACK = 0x6c
	OP_IFDUP        = 0x73
	OP_DEPTH        = 0x74
	OP_DROP         = 0x75
	OP_DUP          = 0x76
	OP_NIP          = 0x77
	OP_OVER         = 0x78
	OP_PICK         = 0x79
	OP_ROLL         = 0x7a
	OP_ROT          = 0x7b
	OP_SWAP         = 0x7c
	OP_TUCK         = 0x7d
	OP_2DROP        = 0x6d
	OP_2DUP         = 0x6e
	OP_3DUP         = 0x6f
	OP_2OVER        = 0x70
	OP_2ROT         = 0x71
	OP_2SWAP        = 0x72

	// Splice
	OP_SIZE = 0x82

	// Bitwise logic
	OP_EQUAL       = 0x87
	OP_EQUALVERIFY = 0x88

	// Arithmetic
	OP_1ADD               = 0x8b
	OP_1SUB               = 0x8c
	OP_NEGATE             = 0x8f
	OP_ABS                = 0x90
	OP_NOT                = 0x91
	OP_0NOTEQUAL          = 0x92
	OP_ADD                = 0x93
	OP_SUB                = 0x94
	OP_BOOLAND            = 0x9a
	OP_BOOLOR             = 0x9b
	OP_NUMEQUAL           = 0x9c
	OP_NUMEQUALVERIFY     = 0x9d
	OP_NUMNOTEQUAL        = 0x9e
	OP_LESSTHAN           = 0x9f
	OP_GREATERTHAN        = 0xa0
	OP_LESSTHANOREQUAL    = 0xa1
	OP_GREATERTHANOREQUAL = 0xa2
	OP_MIN                = 0xa3
	OP_MAX                = 0xa4
	OP_WITHIN             = 0xa5

	// Crypto
	OP_RIPEMD160           = 0xa6
	OP_SHA1                = 0xa7
	OP_SHA256              = 0xa8
	OP_HASH160             = 0xa9
	OP_HASH256             = 0xaa
	OP_CODESEPARATOR       = 0xab
	OP_CHECKSIG            = 0xac
	OP_CHECKSIGVERIFY      = 0xad
	OP_CHECKMULTISIG       = 0xae
	OP_CHECKMULTISIGVERIFY = 0xaf

	// Pseudo-words
	OP_PUBKEYHASH = 0xfd
	OP_PUBKEY     = 0xfe
)

// OpcodeName returns the name of an opcode
func OpcodeName(op byte) string {
	names := map[byte]string{
		OP_0:              "OP_0",
		OP_PUSHDATA1:      "OP_PUSHDATA1",
		OP_PUSHDATA2:      "OP_PUSHDATA2",
		OP_PUSHDATA4:      "OP_PUSHDATA4",
		OP_1NEGATE:        "OP_1NEGATE",
		OP_1:              "OP_1",
		OP_2:              "OP_2",
		OP_3:              "OP_3",
		OP_4:              "OP_4",
		OP_5:              "OP_5",
		OP_6:              "OP_6",
		OP_7:              "OP_7",
		OP_8:              "OP_8",
		OP_9:              "OP_9",
		OP_10:             "OP_10",
		OP_11:             "OP_11",
		OP_12:             "OP_12",
		OP_13:             "OP_13",
		OP_14:             "OP_14",
		OP_15:             "OP_15",
		OP_16:             "OP_16",
		OP_NOP:            "OP_NOP",
		OP_VERIFY:         "OP_VERIFY",
		OP_RETURN:         "OP_RETURN",
		OP_DUP:            "OP_DUP",
		OP_EQUAL:          "OP_EQUAL",
		OP_EQUALVERIFY:    "OP_EQUALVERIFY",
		OP_HASH160:        "OP_HASH160",
		OP_CHECKSIG:       "OP_CHECKSIG",
		OP_CHECKSIGVERIFY: "OP_CHECKSIGVERIFY",
	}

	if name, ok := names[op]; ok {
		return name
	}

	return fmt.Sprintf("OP_UNKNOWN(0x%02x)", op)
}

// IsSmallInt checks if opcode is a small integer (OP_1 to OP_16)
func IsSmallInt(op byte) bool {
	return op >= OP_1 && op <= OP_16
}

// SmallIntValue returns the integer value of a small int opcode
func SmallIntValue(op byte) int {
	if op == OP_0 {
		return 0
	}
	if op == OP_1NEGATE {
		return -1
	}
	if IsSmallInt(op) {
		return int(op - OP_1 + 1)
	}
	return 0
}
