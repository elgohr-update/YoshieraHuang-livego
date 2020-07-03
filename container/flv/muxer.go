package flv

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/utils/pio"
	"github.com/gwuhaolin/livego/utils/uid"

	log "github.com/sirupsen/logrus"
)

var (
	flvHeader = []byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
)

/*
func NewFlv(handler av.Handler, info av.Info) {
	patths := strings.SplitN(info.Key, "/", 2)

	if len(patths) != 2 {
		log.Warning("invalid info")
		return
	}

	w, err := os.OpenFile(*flvFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Error("open file error: ", err)
	}

	writer := NewFLVWriter(patths[0], patths[1], info.URL, w)

	handler.HandleWriter(writer)

	writer.Wait()
	// close flv file
	log.Debug("close flv file")
	writer.ctx.Close()
}
*/

const (
	headerLen = 11
)

// Writer is a writer for flv media
type Writer struct {
	av.RWBaser

	uid    string
	app    string
	title  string
	url    string
	buf    []byte
	closed chan struct{}
	ctx    *os.File
}

// NewWriter returns a writer
func NewWriter(app, title, url string, ctx *os.File) *Writer {
	ret := &Writer{
		RWBaser: av.NewRWBase(time.Second * 10),

		uid:    uid.NewID(),
		app:    app,
		title:  title,
		url:    url,
		ctx:    ctx,
		closed: make(chan struct{}),
		buf:    make([]byte, headerLen),
	}

	ret.ctx.Write(flvHeader)
	pio.PutI32BE(ret.buf[:4], 0)
	ret.ctx.Write(ret.buf[:4])

	return ret
}

// Write write packet into writer
func (writer *Writer) Write(p *av.Packet) error {
	writer.SetPreTime()
	h := writer.buf[:headerLen]
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
	timestamp += writer.BaseTimestamp()
	writer.RecTimestamp(timestamp, uint32(typeID))

	preDataLen := dataLen + headerLen
	timestampbase := timestamp & 0xffffff
	timestampExt := timestamp >> 24 & 0xff

	pio.PutU8(h[0:1], uint8(typeID))
	pio.PutI24BE(h[1:4], int32(dataLen))
	pio.PutI24BE(h[4:7], int32(timestampbase))
	pio.PutU8(h[7:8], uint8(timestampExt))

	if _, err := writer.ctx.Write(h); err != nil {
		return err
	}

	if _, err := writer.ctx.Write(p.Data); err != nil {
		return err
	}

	pio.PutI32BE(h[:4], int32(preDataLen))
	if _, err := writer.ctx.Write(h[:4]); err != nil {
		return err
	}

	return nil
}

// Wait waits for closing
func (writer *Writer) Wait() {
	select {
	case <-writer.closed:
		return
	}
}

// Close close the writer
func (writer *Writer) Close(error) {
	writer.ctx.Close()
	close(writer.closed)
}

// Info return the info
func (writer *Writer) Info() (ret av.Info) {
	ret.UID = writer.uid
	ret.URL = writer.url
	ret.Key = writer.app + "/" + writer.title
	return
}

// Dvr is a dvr
type Dvr struct{}

// Writer get writer from Dvr
func (f *Dvr) Writer(info av.Info) av.WriteCloser {
	paths := strings.SplitN(info.Key, "/", 2)
	if len(paths) != 2 {
		log.Warning("invalid info")
		return nil
	}

	flvDir := configure.Config.GetString("flv_dir")

	err := os.MkdirAll(path.Join(flvDir, paths[0]), 0755)
	if err != nil {
		log.Error("mkdir error: ", err)
		return nil
	}

	fileName := fmt.Sprintf("%s_%d.%s", path.Join(flvDir, info.Key), time.Now().Unix(), "flv")
	log.Debug("flv dvr save stream to: ", fileName)
	w, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Error("open file error: ", err)
		return nil
	}

	writer := NewWriter(paths[0], paths[1], info.URL, w)
	log.Debug("new flv dvr: ", writer.Info())
	return writer
}
