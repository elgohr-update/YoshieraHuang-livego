package rtmprelay

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/protocol/rtmp/core"

	log "github.com/sirupsen/logrus"
)

var (
	stopCtrl = "RTMPRELAY_STOP"
)

// RtmpRelay is a relay for rtmp stream
type RtmpRelay struct {
	PlayURL              string
	PublishURL           string
	csChan               chan core.ChunkStream
	sndctrlChan          chan string
	connectPlayClient    *core.ConnClient
	connectPublishClient *core.ConnClient
	startflag            bool
}

// NewRtmpRelay returns a RtmpRelay
func NewRtmpRelay(playURL *string, publishURL *string) *RtmpRelay {
	return &RtmpRelay{
		PlayURL:              *playURL,
		PublishURL:           *publishURL,
		csChan:               make(chan core.ChunkStream, 500),
		sndctrlChan:          make(chan string),
		connectPlayClient:    nil,
		connectPublishClient: nil,
		startflag:            false,
	}
}

// rcvPlayChunkStream received play chunk stream
func (relay *RtmpRelay) rcvPlayChunkStream() {
	log.Debug("rcvPlayRtmpMediaPacket connectClient.Read...")
	for {
		var rc core.ChunkStream

		if relay.startflag == false {
			relay.connectPlayClient.Close(nil)
			log.Debugf("rcvPlayChunkStream close: playurl=%s, publishurl=%s", relay.PlayURL, relay.PublishURL)
			break
		}
		err := relay.connectPlayClient.Read(&rc)

		if err != nil && err == io.EOF {
			break
		}
		//log.Debugf("connectPlayClient.Read return rc.TypeID=%v length=%d, err=%v", rc.TypeID, len(rc.Data), err)
		switch rc.TypeID {
		case 20, 17:
			r := bytes.NewReader(rc.Data)
			vs, err := relay.connectPlayClient.DecodeBatch(r, amf.AMF0)

			log.Debugf("rcvPlayRtmpMediaPacket: vs=%v, err=%v", vs, err)
		case 18:
			log.Debug("rcvPlayRtmpMediaPacket: metadata....")
		case 8, 9:
			relay.csChan <- rc
		}
	}
}

// sendPublishChunkStream sends publish chunk stream
func (relay *RtmpRelay) sendPublishChunkStream() {
	for {
		select {
		case rc := <-relay.csChan:
			//log.Debugf("sendPublishChunkStream: rc.TypeID=%v length=%d", rc.TypeID, len(rc.Data))
			relay.connectPublishClient.Write(rc)
		case ctrlcmd := <-relay.sndctrlChan:
			if ctrlcmd == stopCtrl {
				relay.connectPublishClient.Close(nil)
				log.Debugf("sendPublishChunkStream close: playurl=%s, publishurl=%s", relay.PlayURL, relay.PublishURL)
				break
			}
		}
	}
}

// Start start the relay
func (relay *RtmpRelay) Start() error {
	if relay.startflag {
		return fmt.Errorf("The rtmprelay already started, playurl=%s, publishurl=%s", relay.PlayURL, relay.PublishURL)
	}

	relay.connectPlayClient = core.NewConnClient()
	relay.connectPublishClient = core.NewConnClient()

	log.Debugf("play server addr:%v starting....", relay.PlayURL)
	err := relay.connectPlayClient.Start(relay.PlayURL, "play")
	if err != nil {
		log.Debugf("connectPlayClient.Start url=%v error", relay.PlayURL)
		return err
	}

	log.Debugf("publish server addr:%v starting....", relay.PublishURL)
	err = relay.connectPublishClient.Start(relay.PublishURL, "publish")
	if err != nil {
		log.Debugf("connectPublishClient.Start url=%v error", relay.PublishURL)
		relay.connectPlayClient.Close(nil)
		return err
	}

	relay.startflag = true
	go relay.rcvPlayChunkStream()
	go relay.sendPublishChunkStream()

	return nil
}

// Stop stops the relay
func (relay *RtmpRelay) Stop() {
	if !relay.startflag {
		log.Debugf("The rtmprelay already stoped, playurl=%s, publishurl=%s", relay.PlayURL, relay.PublishURL)
		return
	}

	relay.startflag = false
	relay.sndctrlChan <- stopCtrl
}
