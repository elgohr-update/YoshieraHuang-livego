package cache

import (
	"bytes"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"

	log "github.com/sirupsen/logrus"
)

const (
	// setDataFrame is the frame of set data
	setDataFrame string = "@setDataFrame"
	// onMetaData is the frame of on metadata
	onMetaData string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, amf.AMF0, setDataFrame); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

// SpecialCache is a cache that can only contain one packet
type SpecialCache struct {
	full bool
	p    *av.Packet
}

// NewSpecialCache returns a SpecialCache
func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

// Write write packet
func (specialCache *SpecialCache) Write(p *av.Packet) {
	specialCache.p = p
	specialCache.full = true
}

// Send send packet to WriteCloser
func (specialCache *SpecialCache) Send(w av.WriteCloser) error {
	if !specialCache.full {
		return nil
	}
	return w.Write(specialCache.p)
}
