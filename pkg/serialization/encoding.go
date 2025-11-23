// This is the MOST IMPORTANT part. Bitcoin's serialization format is byte-perfect.
package serialization

import (
	"encoding/binary"
	"io"
)

// WriteUint32 writes uint32 in little-endian
func WriteUint32(w io.Writer, v uint32) error {
	return binary.Write(w, binary.LittleEndian, v)
}

// WriteInt32 writes int32 in little-endian
func WriteInt32(w io.Writer, v int32) error {
	return binary.Write(w, binary.LittleEndian, v)
}

// WriteUint64 writes uint64 in little-endian
func WriteUint64(w io.Writer, v uint64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

// WriteVarInt writes Bitcoin's compact size format
// This saves space for small numbers
func WriteVarInt(w io.Writer, v uint64) error {
	switch {
	case v < 0xFD: // 0-252: use 1 byte
		_, err := w.Write([]byte{byte(v)})
		return err

	case v <= 0xFFFF: // 253-65535: use 3 bytes
		if _, err := w.Write([]byte{0xFD}); err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, uint16(v))

	case v <= 0xFFFFFFFF: // use 5 bytes
		if _, err := w.Write([]byte{0xFE}); err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, uint32(v))

	default: // use 9 bytes
		if _, err := w.Write([]byte{0xFF}); err != nil {
			return err
		}
		return binary.Write(w, binary.LittleEndian, v)
	}
}

// WriteBytes writes byte slice with length prefix
func WriteBytes(w io.Writer, data []byte) error {
	if err := WriteVarInt(w, uint64(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

// ReadUint32 reads uint32 in little-endian
func ReadUint32(r io.Reader) (uint32, error) {
	var v uint32
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadInt32 reads int32 in little-endian
func ReadInt32(r io.Reader) (int32, error) {
	var v int32
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadUint64 reads uint64 in little-endian
func ReadUint64(r io.Reader) (uint64, error) {
	var v uint64
	err := binary.Read(r, binary.LittleEndian, &v)
	return v, err
}

// ReadVarInt reads Bitcoin's compact size format
func ReadVarInt(r io.Reader) (uint64, error) {
	var firstByte [1]byte
	if _, err := r.Read(firstByte[:]); err != nil {
		return 0, err
	}

	switch firstByte[0] {
	case 0xFD:
		var v uint16
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return 0, err
		}
		return uint64(v), nil
	case 0xFE:
		var v uint32
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return 0, err
		}
		return uint64(v), nil
	case 0xFF:
		var v uint64
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return 0, err
		}
		return v, nil
	default:
		return uint64(firstByte[0]), nil
	}
}

// ReadBytes reads byte slice with length prefix
func ReadBytes(r io.Reader) ([]byte, error) {
	length, err := ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}
	return data, nil
}

/*
```
**VarInt examples:**
```
Value 10:     [0x0A]                    (1 byte)
Value 500:    [0xFD, 0xF4, 0x01]        (3 bytes: 0xFD + little-endian 500)
Value 100000: [0xFE, 0xA0, 0x86, 0x01, 0x00] (5 bytes)
*/
