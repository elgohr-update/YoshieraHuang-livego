package flv

import (
	"fmt"

	"github.com/gwuhaolin/livego/av"
)

var (
	// ErrAvcEndSEQ means error of avc end sequence
	ErrAvcEndSEQ = fmt.Errorf("avc end sequence")
)

// Demuxer does demux things
type Demuxer interface {
	// DemuxH demuxes packet header
	DemuxH(p *av.Packet) error
	// Demux demuxes packet
	Demux(p *av.Packet) error
}

type demuxer struct {
}

// NewDemuxer return a Demuxer
func NewDemuxer() Demuxer {
	return &demuxer{}
}

// DemuxH demuxes packet header
func (d *demuxer) DemuxH(p *av.Packet) error {
	var tag Tag
	_, err := tag.ParseMediaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag

	return nil
}

// Demux demuxes packet
func (d *demuxer) Demux(p *av.Packet) error {
	var tag Tag
	n, err := tag.ParseMediaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	if tag.CodecID() == av.VideoH264 &&
		p.Data[0] == 0x17 && p.Data[1] == 0x02 {
		return ErrAvcEndSEQ
	}
	p.Header = &tag
	p.Data = p.Data[n:]

	return nil
}
