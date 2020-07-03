package cache

import (
	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
)

// Cache is a cache of rtmp
type Cache struct {
	gop      *GopCache
	videoSeq *SpecialCache
	audioSeq *SpecialCache
	metadata *SpecialCache
}

// NewCache returns a Cache
func NewCache() *Cache {
	return &Cache{
		gop:      NewGopCache(configure.Config.GetInt("gop_num")),
		videoSeq: NewSpecialCache(),
		audioSeq: NewSpecialCache(),
		metadata: NewSpecialCache(),
	}
}

// Write writes packet
func (cache *Cache) Write(p av.Packet) {
	if p.IsMetadata {
		cache.metadata.Write(&p)
		return
	}

	if !p.IsVideo {
		ah, ok := p.Header.(av.AudioPacketHeader)
		if ok {
			if ah.SoundFormat() == av.SoundAAC &&
				ah.AACPacketType() == av.AACSeqHeader {
				cache.audioSeq.Write(&p)
			}
			return
		}
	} else {
		vh, ok := p.Header.(av.VideoPacketHeader)
		if ok {
			if vh.IsSeq() {
				cache.videoSeq.Write(&p)
			}
			return
		}
	}

	cache.gop.Write(&p)
}

// Send send the packets to WriteCloser
func (cache *Cache) Send(w av.WriteCloser) error {
	if err := cache.metadata.Send(w); err != nil {
		return err
	}

	if err := cache.videoSeq.Send(w); err != nil {
		return err
	}

	if err := cache.audioSeq.Send(w); err != nil {
		return err
	}

	if err := cache.gop.Send(w); err != nil {
		return err
	}

	return nil
}
