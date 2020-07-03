package amf

import "io"

// ExternalHandler handles
type ExternalHandler func(*Decoder, io.Reader) (interface{}, error)

// Decoder decodes the reader
type Decoder struct {
	refCache         []interface{}
	stringRefs       []string
	objectRefs       []interface{}
	traitRefs        []Trait
	externalHandlers map[string]ExternalHandler
}

// NewDecoder return a decoder
func NewDecoder() *Decoder {
	return &Decoder{
		externalHandlers: make(map[string]ExternalHandler),
	}
}

// RegisterExternalHandler register an ExternalHandler into Decoder
func (d *Decoder) RegisterExternalHandler(name string, f ExternalHandler) {
	d.externalHandlers[name] = f
}

// Trait is trait
type Trait struct {
	Type           string
	Externalizable bool
	Dynamic        bool
	Properties     []string
}

// NewTrait returns a trait
func NewTrait() *Trait {
	return &Trait{}
}
