package httpflv

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/utils/pio"
	"github.com/gwuhaolin/livego/utils/uid"

	log "github.com/sirupsen/logrus"
)

const (
	headerLen   = 11
	maxQueueNum = 1024
)

// Writer is a http flv writer
type Writer struct {
	av.RWBaser

	uid             string
	app, title, url string
	buf             []byte
	closed          bool
	closedChan      chan struct{}
	ctx             http.ResponseWriter
	packetQueue     chan *av.Packet
}

// NewWriter returns a FLV writer
func NewWriter(app, title, url string, ctx http.ResponseWriter) *Writer {
	ret := &Writer{
		RWBaser: av.NewRWBase(time.Second * 10),

		uid:         uid.NewID(),
		app:         app,
		title:       title,
		url:         url,
		ctx:         ctx,
		closedChan:  make(chan struct{}),
		buf:         make([]byte, headerLen),
		packetQueue: make(chan *av.Packet, maxQueueNum),
	}

	ret.ctx.Write([]byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09})
	pio.PutI32BE(ret.buf[:4], 0)
	ret.ctx.Write(ret.buf[:4])
	go func() {
		err := ret.SendPacket()
		if err != nil {
			log.Error("SendPacket error: ", err)
			ret.closed = true
		}
	}()
	return ret
}

// DropPacket drops packets due to queue max
func (flvWriter *Writer) DropPacket(pktQue chan *av.Packet, info av.Info) {
	log.Warningf("[%v] packet queue max!!!", info)
	for i := 0; i < maxQueueNum-84; i++ {
		tmpPkt, ok := <-pktQue
		if ok && tmpPkt.IsVideo {
			videoPkt, ok := tmpPkt.Header.(av.VideoPacketHeader)
			// dont't drop sps config and dont't drop key frame
			if ok && (videoPkt.IsSeq() || videoPkt.IsKeyFrame()) {
				log.Debug("insert keyframe to queue")
				pktQue <- tmpPkt
			}

			if len(pktQue) > maxQueueNum-10 {
				<-pktQue
			}
			// drop other packet
			<-pktQue
		}
		// try to don't drop audio
		if ok && tmpPkt.IsAudio {
			log.Debug("insert audio to queue")
			pktQue <- tmpPkt
		}
	}
	log.Debug("packet queue len: ", len(pktQue))
}

// Write writes packet
func (flvWriter *Writer) Write(p *av.Packet) (err error) {
	err = nil
	if flvWriter.closed {
		err = fmt.Errorf("flvwrite source closed")
		return
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("FLVWriter has already been closed:%v", e)
		}
	}()

	if len(flvWriter.packetQueue) >= maxQueueNum-24 {
		flvWriter.DropPacket(flvWriter.packetQueue, flvWriter.Info())
	} else {
		flvWriter.packetQueue <- p
	}

	return
}

// SendPacket sends packet
func (flvWriter *Writer) SendPacket() error {
	for {
		p, ok := <-flvWriter.packetQueue
		if ok {
			flvWriter.SetPreTime()
			h := flvWriter.buf[:headerLen]
			typeID := av.TagVideo
			if !p.IsVideo {
				if p.IsMetadata {
					var err error
					typeID = av.TagScriptDataAMF0
					p.Data, err = amf.MetaDataReform(p.Data, amf.DEL)
					if err != nil {
						return err
					}
				} else {
					typeID = av.TagAudio
				}
			}
			dataLen := len(p.Data)
			timestamp := p.TimeStamp
			timestamp += flvWriter.BaseTimestamp()
			flvWriter.RecTimestamp(timestamp, uint32(typeID))

			preDataLen := dataLen + headerLen
			timestampbase := timestamp & 0xffffff
			timestampExt := timestamp >> 24 & 0xff

			pio.PutU8(h[0:1], uint8(typeID))
			pio.PutI24BE(h[1:4], int32(dataLen))
			pio.PutI24BE(h[4:7], int32(timestampbase))
			pio.PutU8(h[7:8], uint8(timestampExt))

			if _, err := flvWriter.ctx.Write(h); err != nil {
				return err
			}

			if _, err := flvWriter.ctx.Write(p.Data); err != nil {
				return err
			}

			pio.PutI32BE(h[:4], int32(preDataLen))
			if _, err := flvWriter.ctx.Write(h[:4]); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("closed")
		}

	}
}

// Wait waits for writer closing
func (flvWriter *Writer) Wait() {
	select {
	case <-flvWriter.closedChan:
		return
	}
}

// Close closes the writer
func (flvWriter *Writer) Close(error) {
	log.Debug("http flv closed")
	if !flvWriter.closed {
		close(flvWriter.packetQueue)
		close(flvWriter.closedChan)
	}
	flvWriter.closed = true
}

// Info returns the information
func (flvWriter *Writer) Info() (ret av.Info) {
	ret.UID = flvWriter.uid
	ret.URL = flvWriter.url
	ret.Key = flvWriter.app + "/" + flvWriter.title
	ret.Inter = true
	return
}
