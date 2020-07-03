package amf

import (
	"encoding/json"
	"fmt"
	"io"
)

// DumpBytes dumps bytes to stdout
func DumpBytes(label string, buf []byte, size int) {
	fmt.Printf("Dumping %s (%d bytes):\n", label, size)
	for i := 0; i < size; i++ {
		fmt.Printf("0x%02x ", buf[i])
	}
	fmt.Printf("\n")
}

// Dump dumps things to stdout
func Dump(label string, val interface{}) error {
	json, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return fmt.Errorf("Error dumping %s: %s", label, err)
	}

	fmt.Printf("Dumping %s:\n%s\n", label, json)
	return nil
}

// WriteByte write byte to writer
func WriteByte(w io.Writer, b byte) (err error) {
	bytes := make([]byte, 1)
	bytes[0] = b

	_, err = WriteBytes(w, bytes)

	return
}

// WriteBytes write byte array to writer
func WriteBytes(w io.Writer, bytes []byte) (int, error) {
	return w.Write(bytes)
}

// ReadByte read byte from reader
func ReadByte(r io.Reader) (byte, error) {
	bytes, err := ReadBytes(r, 1)
	if err != nil {
		return 0x00, err
	}

	return bytes[0], nil
}

// ReadBytes read byte array from reader
func ReadBytes(r io.Reader, n int) ([]byte, error) {
	bytes := make([]byte, n)

	m, err := r.Read(bytes)
	if err != nil {
		return bytes, err
	}

	if m != n {
		return bytes, fmt.Errorf("decode read bytes failed: expected %d got %d", m, n)
	}

	return bytes, nil
}

// WriteMarker write marker to writer
func WriteMarker(w io.Writer, m byte) error {
	return WriteByte(w, m)
}

// ReadMarker read the marker from reader
func ReadMarker(r io.Reader) (byte, error) {
	return ReadByte(r)
}

// AssertMarker assert marker to be asserted one
func AssertMarker(r io.Reader, checkMarker bool, m byte) error {
	if checkMarker == false {
		return nil
	}

	marker, err := ReadMarker(r)
	if err != nil {
		return err
	}

	if marker != m {
		return fmt.Errorf("decode assert marker failed: expected %v got %v", m, marker)
	}

	return nil
}
