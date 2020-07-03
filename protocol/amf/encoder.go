package amf

// Encoder is encoder
type Encoder struct {
}

// Version is version
type Version uint8

// Array is a collection of any
type Array []interface{}

// Object is a map
type Object map[string]interface{}

// TypedObject is a map with type of value
type TypedObject struct {
	Type   string
	Object Object
}

// NewTypedObject returns a TypedObject
func NewTypedObject() *TypedObject {
	return &TypedObject{
		Type:   "",
		Object: make(Object),
	}
}
