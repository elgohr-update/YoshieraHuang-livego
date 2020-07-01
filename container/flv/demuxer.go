package flv

import (
	"fmt"

	"github.com/gwuhaolin/livego/av"
)

var (
	ErrAvcEndSEQ = fmt.Errorf("avc end sequence")
)

type Demuxer interface {
	DemuxH(p *av.Packet) error
	Demux(p *av.Packet) error
}

type demuxer struct {
}

func NewDemuxer() Demuxer {
	return &demuxer{}
}

func (d *demuxer) DemuxH(p *av.Packet) error {
	var tag Tag
	_, err := tag.ParseMediaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	p.Header = &tag

	return nil
}

func (d *demuxer) Demux(p *av.Packet) error {
	var tag Tag
	n, err := tag.ParseMediaTagHeader(p.Data, p.IsVideo)
	if err != nil {
		return err
	}
	if tag.CodecID() == av.VIDEO_H264 &&
		p.Data[0] == 0x17 && p.Data[1] == 0x02 {
		return ErrAvcEndSEQ
	}
	p.Header = &tag
	p.Data = p.Data[n:]

	return nil
}
