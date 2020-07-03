package rtmp

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/gwuhaolin/livego/utils/uid"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/container/flv"
	"github.com/gwuhaolin/livego/protocol/rtmp/core"

	log "github.com/sirupsen/logrus"
)

const (
	maxQueueNum         = 1024
	saveStaticsInterval = 5000
)

var (
	readTimeout  = configure.Config.GetInt("read_timeout")
	writeTimeout = configure.Config.GetInt("write_timeout")
)

// Client is the rtmp client
type Client struct {
	handler av.Handler
	getter  av.GetWriter
}

// NewClient return a Client
func NewClient(h av.Handler, getter av.GetWriter) *Client {
	return &Client{
		handler: h,
		getter:  getter,
	}
}

// Dial dials url with specific method
func (c *Client) Dial(url string, method string) error {
	connClient := core.NewConnClient()
	if err := connClient.Start(url, method); err != nil {
		return err
	}
	if method == av.PUBLISH {
		writer := NewVirWriter(connClient)
		log.Debugf("client Dial call NewVirWriter url=%s, method=%s", url, method)
		c.handler.HandleWriter(writer)
	} else if method == av.PLAY {
		reader := NewVirReader(connClient)
		log.Debugf("client Dial call NewVirReader url=%s, method=%s", url, method)
		c.handler.HandleReader(reader)
		if c.getter != nil {
			writer := c.getter.Writer(reader.Info())
			c.handler.HandleWriter(writer)
		}
	}
	return nil
}

// Handler returns a handler
func (c *Client) Handler() av.Handler {
	return c.handler
}

// Server is a rtmp server
type Server struct {
	handler av.Handler
	getter  av.GetWriter
}

// NewServer returns a Server
func NewServer(h av.Handler, getter av.GetWriter) *Server {
	return &Server{
		handler: h,
		getter:  getter,
	}
}

// Serve serves http requests
func (s *Server) Serve(listener net.Listener) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("rtmp serve panic: ", r)
		}
	}()

	for {
		var netconn net.Conn
		netconn, err = listener.Accept()
		if err != nil {
			return
		}
		conn := core.NewConn(netconn, 4*1024)
		log.Debug("new client, connect remote: ", conn.RemoteAddr().String(),
			"local:", conn.LocalAddr().String())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn *core.Conn) error {
	if err := conn.HandshakeServer(); err != nil {
		conn.Close()
		log.Error("handleConn HandshakeServer err: ", err)
		return err
	}
	connServer := core.NewConnServer(conn)

	if err := connServer.ReadMsg(); err != nil {
		conn.Close()
		log.Error("handleConn read msg err: ", err)
		return err
	}

	appname, name, _ := connServer.GetInfo()

	if ret := configure.CheckAppName(appname); !ret {
		err := fmt.Errorf("application name=%s is not configured", appname)
		conn.Close()
		log.Error("CheckAppName err: ", err)
		return err
	}

	log.Debugf("handleConn: IsPublisher=%v", connServer.IsPublisher())
	if connServer.IsPublisher() {
		channel, err := configure.RoomKeys.GetChannel(name)
		if err != nil {
			err := fmt.Errorf("invalid key")
			conn.Close()
			log.Error("CheckKey err: ", err)
			return err
		}
		connServer.PublishInfo.Name = channel
		if pushlist, ret := configure.GetStaticPushURLList(appname); ret && (pushlist != nil) {
			log.Debugf("GetStaticPushUrlList: %v", pushlist)
		}
		reader := NewVirReader(connServer)
		s.handler.HandleReader(reader)
		log.Debugf("new publisher: %+v", reader.Info())

		if s.getter != nil {
			writeType := reflect.TypeOf(s.getter)
			log.Debugf("handleConn:writeType=%v", writeType)
			writer := s.getter.Writer(reader.Info())
			s.handler.HandleWriter(writer)
		}
		flvWriter := new(flv.Dvr)
		s.handler.HandleWriter(flvWriter.Writer(reader.Info()))
	} else {
		writer := NewVirWriter(connServer)
		log.Debugf("new player: %+v", writer.Info())
		s.handler.HandleWriter(writer)
	}

	return nil
}

// GetInfo returns a struct that can return a info
type GetInfo interface {
	GetInfo() (string, string, string)
}

// StreamReadWriteCloser is a ReadWriteCloser for Stream
type StreamReadWriteCloser interface {
	GetInfo
	Close(error)
	Write(core.ChunkStream) error
	Read(*core.ChunkStream) error
}

// StaticsBW is static bw
type StaticsBW struct {
	StreamID               uint32
	VideoDatainBytes       uint64
	LastVideoDatainBytes   uint64
	VideoSpeedInBytesperMS uint64

	AudioDatainBytes       uint64
	LastAudioDatainBytes   uint64
	AudioSpeedInBytesperMS uint64

	LastTimestamp int64
}

// VirWriter is a writer for vir
type VirWriter struct {
	av.RWBaser

	uid         string
	closed      bool
	conn        StreamReadWriteCloser
	packetQueue chan *av.Packet
	WriteBWInfo StaticsBW
}

// NewVirWriter return a VirWriter
func NewVirWriter(conn StreamReadWriteCloser) *VirWriter {
	ret := &VirWriter{
		RWBaser: av.NewRWBase(time.Second * time.Duration(writeTimeout)),

		uid:         uid.NewID(),
		conn:        conn,
		packetQueue: make(chan *av.Packet, maxQueueNum),
		WriteBWInfo: StaticsBW{0, 0, 0, 0, 0, 0, 0, 0},
	}

	go ret.Check()
	go func() {
		err := ret.SendPacket()
		if err != nil {
			log.Warning(err)
		}
	}()
	return ret
}

// SaveStatics save the statics
func (v *VirWriter) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.WriteBWInfo.StreamID = streamid
	if isVideoFlag {
		v.WriteBWInfo.VideoDatainBytes = v.WriteBWInfo.VideoDatainBytes + length
	} else {
		v.WriteBWInfo.AudioDatainBytes = v.WriteBWInfo.AudioDatainBytes + length
	}

	if v.WriteBWInfo.LastTimestamp == 0 {
		v.WriteBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.WriteBWInfo.LastTimestamp) >= saveStaticsInterval {
		diffTimestamp := (nowInMS - v.WriteBWInfo.LastTimestamp) / 1000

		v.WriteBWInfo.VideoSpeedInBytesperMS = (v.WriteBWInfo.VideoDatainBytes - v.WriteBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.WriteBWInfo.AudioSpeedInBytesperMS = (v.WriteBWInfo.AudioDatainBytes - v.WriteBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.WriteBWInfo.LastVideoDatainBytes = v.WriteBWInfo.VideoDatainBytes
		v.WriteBWInfo.LastAudioDatainBytes = v.WriteBWInfo.AudioDatainBytes
		v.WriteBWInfo.LastTimestamp = nowInMS
	}
}

// Check check
func (v *VirWriter) Check() {
	var c core.ChunkStream
	for {
		if err := v.conn.Read(&c); err != nil {
			v.Close(err)
			return
		}
	}
}

// DropPacket drops packet due to queue max
func (v *VirWriter) DropPacket(pktQue chan *av.Packet, info av.Info) {
	log.Warningf("[%v] packet queue max!!!", info)
	for i := 0; i < maxQueueNum-84; i++ {
		tmpPkt, ok := <-pktQue
		// try to don't drop audio
		if ok && tmpPkt.IsAudio {
			if len(pktQue) > maxQueueNum-2 {
				log.Debug("drop audio pkt")
				<-pktQue
			} else {
				pktQue <- tmpPkt
			}

		}

		if ok && tmpPkt.IsVideo {
			videoPkt, ok := tmpPkt.Header.(av.VideoPacketHeader)
			// dont't drop sps config and dont't drop key frame
			if ok && (videoPkt.IsSeq() || videoPkt.IsKeyFrame()) {
				pktQue <- tmpPkt
			}
			if len(pktQue) > maxQueueNum-10 {
				log.Debug("drop video pkt")
				<-pktQue
			}
		}

	}
	log.Debug("packet queue len: ", len(pktQue))
}

// Write writes packet
func (v *VirWriter) Write(p *av.Packet) (err error) {
	err = nil

	if v.closed {
		err = fmt.Errorf("VirWriter closed")
		return
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("VirWriter has already been closed:%v", e)
		}
	}()
	if len(v.packetQueue) >= maxQueueNum-24 {
		v.DropPacket(v.packetQueue, v.Info())
	} else {
		v.packetQueue <- p
	}

	return
}

// SendPacket sends packet
func (v *VirWriter) SendPacket() error {
	Flush := reflect.ValueOf(v.conn).MethodByName("Flush")
	var cs core.ChunkStream
	for {
		p, ok := <-v.packetQueue
		if ok {
			cs.Data = p.Data
			cs.Length = uint32(len(p.Data))
			cs.StreamID = p.StreamID
			cs.Timestamp = p.TimeStamp
			cs.Timestamp += v.BaseTimestamp()

			if p.IsVideo {
				cs.TypeID = av.TagVideo
			} else {
				if p.IsMetadata {
					cs.TypeID = av.TagScriptDataAMF0
				} else {
					cs.TypeID = av.TagAudio
				}
			}

			v.SaveStatics(p.StreamID, uint64(cs.Length), p.IsVideo)
			v.SetPreTime()
			v.RecTimestamp(cs.Timestamp, cs.TypeID)
			err := v.conn.Write(cs)
			if err != nil {
				v.closed = true
				return err
			}
			Flush.Call(nil)
		} else {
			return fmt.Errorf("closed")
		}

	}

}

// Info returns info
func (v *VirWriter) Info() (ret av.Info) {
	ret.UID = v.uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		log.Warning(err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	ret.Inter = true
	return
}

// Close closes this VirWriter
func (v *VirWriter) Close(err error) {
	log.Warning("player ", v.Info(), "closed: "+err.Error())
	if !v.closed {
		close(v.packetQueue)
	}
	v.closed = true
	v.conn.Close(err)
}

// VirReader is a virReader
type VirReader struct {
	av.RWBaser

	uid        string
	demuxer    flv.Demuxer
	conn       StreamReadWriteCloser
	ReadBWInfo StaticsBW
}

// NewVirReader returns a virReader
func NewVirReader(conn StreamReadWriteCloser) *VirReader {
	return &VirReader{
		RWBaser: av.NewRWBase(time.Second * time.Duration(writeTimeout)),

		uid:     uid.NewID(),
		conn:    conn,
		demuxer: flv.NewDemuxer(),
		ReadBWInfo: StaticsBW{
			StreamID:               0,
			VideoDatainBytes:       0,
			LastVideoDatainBytes:   0,
			VideoSpeedInBytesperMS: 0,
			AudioDatainBytes:       0,
			LastAudioDatainBytes:   0,
			AudioSpeedInBytesperMS: 0,
			LastTimestamp:          0,
		},
	}
}

// SaveStatics save the statics
func (v *VirReader) SaveStatics(streamid uint32, length uint64, isVideoFlag bool) {
	nowInMS := int64(time.Now().UnixNano() / 1e6)

	v.ReadBWInfo.StreamID = streamid
	if isVideoFlag {
		v.ReadBWInfo.VideoDatainBytes = v.ReadBWInfo.VideoDatainBytes + length
	} else {
		v.ReadBWInfo.AudioDatainBytes = v.ReadBWInfo.AudioDatainBytes + length
	}

	if v.ReadBWInfo.LastTimestamp == 0 {
		v.ReadBWInfo.LastTimestamp = nowInMS
	} else if (nowInMS - v.ReadBWInfo.LastTimestamp) >= saveStaticsInterval {
		diffTimestamp := (nowInMS - v.ReadBWInfo.LastTimestamp) / 1000

		//log.Printf("now=%d, last=%d, diff=%d", nowInMS, v.ReadBWInfo.LastTimestamp, diffTimestamp)
		v.ReadBWInfo.VideoSpeedInBytesperMS = (v.ReadBWInfo.VideoDatainBytes - v.ReadBWInfo.LastVideoDatainBytes) * 8 / uint64(diffTimestamp) / 1000
		v.ReadBWInfo.AudioSpeedInBytesperMS = (v.ReadBWInfo.AudioDatainBytes - v.ReadBWInfo.LastAudioDatainBytes) * 8 / uint64(diffTimestamp) / 1000

		v.ReadBWInfo.LastVideoDatainBytes = v.ReadBWInfo.VideoDatainBytes
		v.ReadBWInfo.LastAudioDatainBytes = v.ReadBWInfo.AudioDatainBytes
		v.ReadBWInfo.LastTimestamp = nowInMS
	}
}

// Read read to packet
func (v *VirReader) Read(p *av.Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Warning("rtmp read packet panic: ", r)
		}
	}()

	v.SetPreTime()
	var cs core.ChunkStream
	for {
		err = v.conn.Read(&cs)
		if err != nil {
			return err
		}
		if cs.TypeID == av.TagAudio ||
			cs.TypeID == av.TagVideo ||
			cs.TypeID == av.TagScriptDataAMF0 ||
			cs.TypeID == av.TagScriptDataAMF3 {
			break
		}
	}

	p.IsAudio = cs.TypeID == av.TagAudio
	p.IsVideo = cs.TypeID == av.TagVideo
	p.IsMetadata = cs.TypeID == av.TagScriptDataAMF0 || cs.TypeID == av.TagScriptDataAMF3
	p.StreamID = cs.StreamID
	p.Data = cs.Data
	p.TimeStamp = cs.Timestamp

	v.SaveStatics(p.StreamID, uint64(len(p.Data)), p.IsVideo)
	v.demuxer.DemuxH(p)
	return err
}

// Info returns info
func (v *VirReader) Info() (ret av.Info) {
	ret.UID = v.uid
	_, _, URL := v.conn.GetInfo()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		log.Warning(err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	return
}

// Close close this reader
func (v *VirReader) Close(err error) {
	log.Debug("publisher ", v.Info(), "closed: "+err.Error())
	v.conn.Close(err)
}
