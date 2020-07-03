package amf

const (
	// AMF0 is the amf0 marker
	AMF0 = 0x00
	// AMF3 is the amf3 marker
	AMF3 = 0x03
)

// AMF0 Marker
const (
	// AMF0NumberMarker is the number marker for AMF0
	AMF0NumberMarker = 0x00
	// AMF0BooleanMarker is the boolean marker for AMF0
	AMF0BooleanMarker = 0x01
	// AMF0StringMarker is the string marker for AMF0
	AMF0StringMarker = 0x02
	// AMF0ObjectMarker is the object marker for AMF0
	AMF0ObjectMarker = 0x03
	// AMF0MovieclipMarker is the movieclip marker for AMF0
	AMF0MovieclipMarker = 0x04
	// AMF0NullMarker is the null marker for AMF0
	AMF0NullMarker = 0x05
	// AMF0UndefinedMarker is the undefined marker for AMF0
	AMF0UndefinedMarker = 0x06
	// AMF0ReferenceMaker is the reference marker for AMF0
	AMF0ReferenceMarker = 0x07
	// AMF0ECMAArrayMarker is the ECMA array marker for AMF0
	AMF0EcmaArrayMarker = 0x08
	// AMF0ObjectEndMarker is the object end marker for AMF0
	AMF0ObjectEndMarker = 0x09
	// AMF0StrictArrayMarker is the strict array marker for AMF0
	AMF0StrictArrayMarker = 0x0a
	// AMF0DataMarker is the date marker for AMF0
	AMF0DateMarker = 0x0b
	// AMF0LongStringMarker is the long string marker for AMF0
	AMF0LongStringMarker = 0x0c
	// AMF0UnsupportedMarker is the unsupported marker for AMF0
	AMF0UnsupportedMarker = 0x0d
	// AMF0RecordsetMarker is the record set marker for AMF0
	AMF0RecordsetMarker = 0x0e
	// AMF0XMLDocumentMarker is the XML document marker for AMF0
	AMF0XMLDocumentMarker = 0x0f
	// AMF0TypedObjectMarker is the typed object marker for AMF0
	AMF0TypedObjectMarker = 0x10
	// AMF0AcmplusObjectMarker is the acmplus object marker for AMF0
	AMF0AcmplusObjectMarker = 0x11
)

// AMF0 constants
const (
	// AMF0BooleanFalse denotes false in AMF0
	AMF0BooleanFalse = 0x00
	// AMF0BooleanTrue denotes true in AMF0
	AMF0BooleanTrue = 0x01
	// AMF0StringMax denotes max string length
	AMF0StringMax = 65535
)

const (
	// AMF3UndefinedMarker denotes undefined marker in AMF3
	AMF3UndefinedMarker = 0x00
	// AMF3NullMarker denotes null marker in AMF3
	AMF3NullMarker = 0x01
	// AMF3FalseMarker denotes false marker in AMF3
	AMF3FalseMarker = 0x02
	// AMF3TrueMarker denotes true marker in AMF3
	AMF3TrueMarker = 0x03
	// AMF3IntegerMarker denotes integer marker in AMF3
	AMF3IntegerMarker = 0x04
	// AMF3DoubleMarker denotes double marker in AMF3
	AMF3DoubleMarker = 0x05
	// AMF3StringMarker denotes string marker in AMF3
	AMF3StringMarker = 0x06
	// AMF3XMLDocumentMarker denotes XML document marker in AMF3
	AMF3XMLDocumentMarker = 0x07
	// AMF3DateMarker denotes date document marker in AMF3
	AMF3DateMarker = 0x08
	// AMF3ArrayMarker denotes array document marker in AMF3
	AMF3ArrayMarker = 0x09
	// AMF3ObjectMarker denotes object marker in AMF3
	AMF3ObjectMarker = 0x0a
	// AMF3XMLStringMarker denotes XML string marker in AMF3
	AMF3XMLStringMarker = 0x0b
	// AMF3BytearrayMarker denotes byte array marker in AMF3
	AMF3BytearrayMarker = 0x0c
)

// AMF3 constants
const (
	// AMF3IntegerMax denotes max integer in AMF3
	AMF3IntegerMax = 536870911
)
