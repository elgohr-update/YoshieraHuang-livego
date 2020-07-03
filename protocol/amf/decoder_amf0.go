package amf

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DecodeAmf0 decodes amf0 reader.
// actually amf0 polymorphic router
func (d *Decoder) DecodeAmf0(r io.Reader) (interface{}, error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}

	switch marker {
	case AMF0NumberMarker:
		return d.DecodeAmf0Number(r, false)
	case AMF0BooleanMarker:
		return d.DecodeAmf0Boolean(r, false)
	case AMF0StringMarker:
		return d.DecodeAmf0String(r, false)
	case AMF0ObjectMarker:
		return d.DecodeAmf0Object(r, false)
	case AMF0MovieclipMarker:
		return nil, fmt.Errorf("decode amf0: unsupported type movieclip")
	case AMF0NullMarker:
		return d.DecodeAmf0Null(r, false)
	case AMF0UndefinedMarker:
		return d.DecodeAmf0Undefined(r, false)
	case AMF0ReferenceMarker:
		return nil, fmt.Errorf("decode amf0: unsupported type reference")
	case AMF0EcmaArrayMarker:
		return d.DecodeAmf0EcmaArray(r, false)
	case AMF0StrictArrayMarker:
		return d.DecodeAmf0StrictArray(r, false)
	case AMF0DateMarker:
		return d.DecodeAmf0Date(r, false)
	case AMF0LongStringMarker:
		return d.DecodeAmf0LongString(r, false)
	case AMF0UnsupportedMarker:
		return d.DecodeAmf0Unsupported(r, false)
	case AMF0RecordsetMarker:
		return nil, fmt.Errorf("decode amf0: unsupported type recordset")
	case AMF0XMLDocumentMarker:
		return d.DecodeAmf0XmlDocument(r, false)
	case AMF0TypedObjectMarker:
		return d.DecodeAmf0TypedObject(r, false)
	case AMF0AcmplusObjectMarker:
		return d.DecodeAmf3(r)
	}

	return nil, fmt.Errorf("decode amf0: unsupported type %d", marker)
}

// DecodeAmf0Number decodes amf0 number.
// marker: 1 byte 0x00
// format: 8 byte big endian float64
func (d *Decoder) DecodeAmf0Number(r io.Reader, decodeMarker bool) (result float64, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0NumberMarker); err != nil {
		return
	}

	err = binary.Read(r, binary.BigEndian, &result)
	if err != nil {
		return float64(0), fmt.Errorf("amf0 decode: unable to read number: %s", err)
	}

	return
}

// DecodeAmf0Boolean decode amf0 boolean
// marker: 1 byte 0x01
// format: 1 byte, 0x00 = false, 0x01 = true
func (d *Decoder) DecodeAmf0Boolean(r io.Reader, decodeMarker bool) (result bool, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0BooleanMarker); err != nil {
		return
	}

	var b byte
	if b, err = ReadByte(r); err != nil {
		return
	}

	if b == AMF0BooleanFalse {
		return false, nil
	} else if b == AMF0BooleanTrue {
		return true, nil
	}

	return false, fmt.Errorf("decode amf0: unexpected value %v for boolean", b)
}

// DecodeAmf0String decodes amf0 string
// marker: 1 byte 0x02
// format:
// - 2 byte big endian uint16 header to determine size
// - n (size) byte utf8 string
func (d *Decoder) DecodeAmf0String(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0StringMarker); err != nil {
		return
	}

	var length uint16
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode string length: %s", err)
	}

	var bytes = make([]byte, length)
	if bytes, err = ReadBytes(r, int(length)); err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode string value: %s", err)
	}

	return string(bytes), nil
}

// DecodeAmf0Object decodes amf0 object
// marker: 1 byte 0x03
// format:
// - loop encoded string followed by encoded value
// - terminated with empty string followed by 1 byte 0x09
func (d *Decoder) DecodeAmf0Object(r io.Reader, decodeMarker bool) (Object, error) {
	if err := AssertMarker(r, decodeMarker, AMF0ObjectMarker); err != nil {
		return nil, err
	}

	result := make(Object)
	d.refCache = append(d.refCache, result)

	for {
		key, err := d.DecodeAmf0String(r, false)
		if err != nil {
			return nil, err
		}

		if key == "" {
			if err = AssertMarker(r, true, AMF0ObjectEndMarker); err != nil {
				return nil, fmt.Errorf("decode amf0: expected object end marker: %s", err)
			}

			break
		}

		value, err := d.DecodeAmf0(r)
		if err != nil {
			return nil, fmt.Errorf("decode amf0: unable to decode object value: %s", err)
		}

		result[key] = value
	}

	return result, nil

}

// DecodeAmf0Null decodes the amf0 null
// marker: 1 byte 0x05
// no additional data
func (d *Decoder) DecodeAmf0Null(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF0NullMarker)
	return
}

// DecodeAmf0Undefined decodes the amf0 undefined
// marker: 1 byte 0x06
// no additional data
func (d *Decoder) DecodeAmf0Undefined(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF0UndefinedMarker)
	return
}

// DecodeAmf0EcmaArray decodes ecma array of amf0
// marker: 1 byte 0x08
// format:
// - 4 byte big endian uint32 with length of associative array
// - normal object format:
//   - loop encoded string followed by encoded value
//   - terminated with empty string followed by 1 byte 0x09
func (d *Decoder) DecodeAmf0EcmaArray(r io.Reader, decodeMarker bool) (Object, error) {
	if err := AssertMarker(r, decodeMarker, AMF0EcmaArrayMarker); err != nil {
		return nil, err
	}

	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)

	result, err := d.DecodeAmf0Object(r, false)
	if err != nil {
		return nil, fmt.Errorf("decode amf0: unable to decode ecma array object: %s", err)
	}

	return result, nil
}

// DecodeAmf0StrictArray decodes amf0 strict array
// marker: 1 byte 0x0a
// format:
// - 4 byte big endian uint32 to determine length of associative array
// - n (length) encoded values
func (d *Decoder) DecodeAmf0StrictArray(r io.Reader, decodeMarker bool) (result Array, err error) {
	if err := AssertMarker(r, decodeMarker, AMF0StrictArrayMarker); err != nil {
		return nil, err
	}

	var length uint32
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("decode amf0: unable to decode strict array length: %s", err)
	}

	d.refCache = append(d.refCache, result)

	for i := uint32(0); i < length; i++ {
		tmp, err := d.DecodeAmf0(r)
		if err != nil {
			return nil, fmt.Errorf("decode amf0: unable to decode strict array object: %s", err)
		}
		result = append(result, tmp)
	}

	return result, nil
}

// DecodeAmf0Date decodes date of amf0
// marker: 1 byte 0x0b
// format:
// - normal number format:
//   - 8 byte big endian float64
// - 2 byte unused
func (d *Decoder) DecodeAmf0Date(r io.Reader, decodeMarker bool) (result float64, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0DateMarker); err != nil {
		return
	}

	if result, err = d.DecodeAmf0Number(r, false); err != nil {
		return float64(0), fmt.Errorf("decode amf0: unable to decode float in date: %s", err)
	}

	if _, err = ReadBytes(r, 2); err != nil {
		return float64(0), fmt.Errorf("decode amf0: unable to read 2 trail bytes in date: %s", err)
	}

	return
}

// DecodeAmf0LongString decodes long string of amf0
// marker: 1 byte 0x0c
// format:
// - 4 byte big endian uint32 header to determine size
// - n (size) byte utf8 string
func (d *Decoder) DecodeAmf0LongString(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0LongStringMarker); err != nil {
		return
	}

	var length uint32
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode long string length: %s", err)
	}

	var bytes = make([]byte, length)
	if bytes, err = ReadBytes(r, int(length)); err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode long string value: %s", err)
	}

	return string(bytes), nil
}

// DecodeAmf0Unsupported decodes unsupported of amf0
// marker: 1 byte 0x0d
// no additional data
func (d *Decoder) DecodeAmf0Unsupported(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF0UnsupportedMarker)
	return
}

// DecodeAmf0XmlDocument decode XML document of amf0
// marker: 1 byte 0x0f
// format:
// - normal long string format
//   - 4 byte big endian uint32 header to determine size
//   - n (size) byte utf8 string
func (d *Decoder) DecodeAmf0XmlDocument(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0XMLDocumentMarker); err != nil {
		return
	}

	return d.DecodeAmf0LongString(r, false)
}

// DecodeAmf0TypedObject decodes typed object of amf0
// marker: 1 byte 0x10
// format:
// - normal string format:
//   - 2 byte big endian uint16 header to determine size
//   - n (size) byte utf8 string
// - normal object format:
//   - loop encoded string followed by encoded value
//   - terminated with empty string followed by 1 byte 0x09
func (d *Decoder) DecodeAmf0TypedObject(r io.Reader, decodeMarker bool) (TypedObject, error) {
	result := *new(TypedObject)

	err := AssertMarker(r, decodeMarker, AMF0TypedObjectMarker)
	if err != nil {
		return result, err
	}

	d.refCache = append(d.refCache, result)

	result.Type, err = d.DecodeAmf0String(r, false)
	if err != nil {
		return result, fmt.Errorf("decode amf0: typed object unable to determine type: %s", err)
	}

	result.Object, err = d.DecodeAmf0Object(r, false)
	if err != nil {
		return result, fmt.Errorf("decode amf0: typed object unable to determine object: %s", err)
	}

	return result, nil
}
