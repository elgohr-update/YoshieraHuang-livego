package rtmprelay

import (
	"fmt"
	"sync"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/protocol/rtmp/core"

	log "github.com/sirupsen/logrus"
)

// StaticPush is a static push
type StaticPush struct {
	RtmpURL       string
	packetChan    chan *av.Packet
	sndctrlChan   chan string
	connectClient *core.ConnClient
	startflag     bool
}

// NewStaticPush returns a StaticPush
func NewStaticPush(rtmpurl string) *StaticPush {
	return &StaticPush{
		RtmpURL:       rtmpurl,
		packetChan:    make(chan *av.Packet, 500),
		sndctrlChan:   make(chan string),
		connectClient: nil,
		startflag:     false,
	}
}

var staticPushMap = make(map[string](*StaticPush))
var mapLock = new(sync.RWMutex)

var (
	staticRelayStopCtrl = "STATIC_RTMPRELAY_STOP"
)

// GetStaticPushList get the static push list of an application
func GetStaticPushList(appname string) ([]string, error) {
	pushurlList, ok := configure.GetStaticPushURLList(appname)

	if !ok {
		return nil, fmt.Errorf("no static push url")
	}

	return pushurlList, nil
}

// GetAndCreateStaticPushObject get the staticpush of given rtmpurl, or create one
func GetAndCreateStaticPushObject(rtmpurl string) *StaticPush {
	mapLock.RLock()
	staticpush, ok := staticPushMap[rtmpurl]
	log.Debugf("GetAndCreateStaticPushObject: %s, return %v", rtmpurl, ok)
	if !ok {
		mapLock.RUnlock()
		newStaticpush := NewStaticPush(rtmpurl)

		mapLock.Lock()
		staticPushMap[rtmpurl] = newStaticpush
		mapLock.Unlock()

		return newStaticpush
	}
	mapLock.RUnlock()

	return staticpush
}

// GetStaticPushObject gets static push object
func GetStaticPushObject(rtmpurl string) (*StaticPush, error) {
	mapLock.RLock()
	if staticpush, ok := staticPushMap[rtmpurl]; ok {
		mapLock.RUnlock()
		return staticpush, nil
	}
	mapLock.RUnlock()

	return nil, fmt.Errorf("staticPushMap[%s] not exist", rtmpurl)
}

// ReleaseStaticPushObject release one static_push object
func ReleaseStaticPushObject(rtmpurl string) {
	mapLock.RLock()
	if _, ok := staticPushMap[rtmpurl]; ok {
		mapLock.RUnlock()

		log.Debugf("ReleaseStaticPushObject %s ok", rtmpurl)
		mapLock.Lock()
		delete(staticPushMap, rtmpurl)
		mapLock.Unlock()
	} else {
		mapLock.RUnlock()
		log.Debugf("ReleaseStaticPushObject: not find %s", rtmpurl)
	}
}

// Start starts publishing
func (sp *StaticPush) Start() error {
	if sp.startflag {
		return fmt.Errorf("StaticPush already start %s", sp.RtmpURL)
	}

	sp.connectClient = core.NewConnClient()

	log.Debugf("static publish server addr:%v starting....", sp.RtmpURL)
	err := sp.connectClient.Start(sp.RtmpURL, "publish")
	if err != nil {
		log.Debugf("connectClient.Start url=%v error", sp.RtmpURL)
		return err
	}
	log.Debugf("static publish server addr:%v started, streamid=%d", sp.RtmpURL, sp.connectClient.StreamID())
	go sp.Handle()

	sp.startflag = true
	return nil
}

// Stop stops the stop
func (sp *StaticPush) Stop() {
	if !sp.startflag {
		return
	}

	log.Debugf("StaticPush Stop: %s", sp.RtmpURL)
	sp.sndctrlChan <- staticRelayStopCtrl
	sp.startflag = false
}

// Write writes a packet
func (sp *StaticPush) Write(packet *av.Packet) {
	if !sp.startflag {
		return
	}

	sp.packetChan <- packet
}

// Send send packet
func (sp *StaticPush) Send(p *av.Packet) {
	if !sp.startflag {
		return
	}
	var cs core.ChunkStream

	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = sp.connectClient.StreamID()
	cs.Timestamp = p.TimeStamp
	//cs.Timestamp += v.BaseTimeStamp()

	//log.Printf("Static sendPacket: rtmpurl=%s, length=%d, streamid=%d",
	//	self.RtmpUrl, len(p.Data), cs.StreamID)
	if p.IsVideo {
		cs.TypeID = av.TagVideo
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TagScriptDataAMF0
		} else {
			cs.TypeID = av.TagAudio
		}
	}

	sp.connectClient.Write(cs)
}

// Handle handles packet
func (sp *StaticPush) Handle() {
	if !sp.IsStart() {
		log.Debugf("static push %s not started", sp.RtmpURL)
		return
	}

	for {
		select {
		case packet := <-sp.packetChan:
			sp.Send(packet)
		case ctrlcmd := <-sp.sndctrlChan:
			if ctrlcmd == staticRelayStopCtrl {
				sp.connectClient.Close(nil)
				log.Debugf("Static HandleAvPacket close: publishurl=%s", sp.RtmpURL)
				break
			}
		}
	}
}

// IsStart returns if this is started
func (sp *StaticPush) IsStart() bool {
	return sp.startflag
}
