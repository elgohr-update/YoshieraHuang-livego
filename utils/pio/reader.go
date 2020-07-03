package pio

// U8 converts byte slice to uint8
func U8(b []byte) (i uint8) {
	return b[0]
}

// U16BE converts big-endian byte slice to uint16
func U16BE(b []byte) (i uint16) {
	i = uint16(b[0])
	i <<= 8
	i |= uint16(b[1])
	return
}

// I16BE converts big-endian byte slice to int16
func I16BE(b []byte) (i int16) {
	i = int16(b[0])
	i <<= 8
	i |= int16(b[1])
	return
}

// I24BE converts big-endian byte slice to int24
func I24BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	return
}

// U24BE converts big-endian byte slice to uint24
func U24BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	return
}

// I32BE converts big-endian byte slice to int32
func I32BE(b []byte) (i int32) {
	i = int32(int8(b[0]))
	i <<= 8
	i |= int32(b[1])
	i <<= 8
	i |= int32(b[2])
	i <<= 8
	i |= int32(b[3])
	return
}

// U32LE converts little-endian byte slice to int32
func U32LE(b []byte) (i uint32) {
	i = uint32(b[3])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[0])
	return
}

// U32BE converts big-endian byte slice to uint32
func U32BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[3])
	return
}

// U40BE converts big-endian byte slice to uint40
func U40BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	return
}

// U64BE converts big-endian byte slice to uint64
func U64BE(b []byte) (i uint64) {
	i = uint64(b[0])
	i <<= 8
	i |= uint64(b[1])
	i <<= 8
	i |= uint64(b[2])
	i <<= 8
	i |= uint64(b[3])
	i <<= 8
	i |= uint64(b[4])
	i <<= 8
	i |= uint64(b[5])
	i <<= 8
	i |= uint64(b[6])
	i <<= 8
	i |= uint64(b[7])
	return
}

// I64BE converts big-endian byte slice to int64
func I64BE(b []byte) (i int64) {
	i = int64(int8(b[0]))
	i <<= 8
	i |= int64(b[1])
	i <<= 8
	i |= int64(b[2])
	i <<= 8
	i |= int64(b[3])
	i <<= 8
	i |= int64(b[4])
	i <<= 8
	i |= int64(b[5])
	i <<= 8
	i |= int64(b[6])
	i <<= 8
	i |= int64(b[7])
	return
}
