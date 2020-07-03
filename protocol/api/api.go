package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/protocol/rtmp"
	"github.com/gwuhaolin/livego/protocol/rtmp/rtmprelay"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

// Response contains ResponseWriter, status and data
type Response struct {
	w      http.ResponseWriter
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

// SendJSON send json data as response with status
func (r *Response) SendJSON() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Status)
	return r.w.Write(resp)
}

// type Operation struct {
// 	Method string `json:"method"`
// 	URL    string `json:"url"`
// 	Stop   bool   `json:"stop"`
// }

// type OperationChange struct {
// 	Method    string `json:"method"`
// 	SourceURL string `json:"source_url"`
// 	TargetURL string `json:"target_url"`
// 	Stop      bool   `json:"stop"`
// }

// type ClientInfo struct {
// 	url              string
// 	rtmpRemoteClient *rtmp.Client
// 	rtmpLocalClient  *rtmp.Client
// }

// Server serve the http api
type Server struct {
	handler  av.Handler
	session  map[string]*rtmprelay.RtmpRelay
	rtmpAddr string
}

// NewServer return a new Server
func NewServer(h av.Handler, rtmpAddr string) *Server {
	return &Server{
		handler:  h,
		session:  make(map[string]*rtmprelay.RtmpRelay),
		rtmpAddr: rtmpAddr,
	}
}

// JWTMiddleware is a jwt middleware
// If jwt.secret is specified in config, this middleware will be activated.
func JWTMiddleware(next http.Handler) http.Handler {
	isJWT := len(configure.Config.GetString("jwt.secret")) > 0
	if !isJWT {
		return next
	}

	log.Info("Using JWT middleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var algorithm jwt.SigningMethod
		if len(configure.Config.GetString("jwt.algorithm")) > 0 {
			algorithm = jwt.GetSigningMethod(configure.Config.GetString("jwt.algorithm"))
		}

		if algorithm == nil {
			algorithm = jwt.SigningMethodHS256
		}

		jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
			Extractor: jwtmiddleware.FromFirst(jwtmiddleware.FromAuthHeader, jwtmiddleware.FromParameter("jwt")),
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte(configure.Config.GetString("jwt.secret")), nil
			},
			SigningMethod: algorithm,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
				res := &Response{
					w:      w,
					Status: 403,
					Data:   err,
				}
				res.SendJSON()
			},
		})

		jwtMiddleware.HandlerWithNext(w, r, next.ServeHTTP)
	})
}

// Serve serves http request
func (s *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()

	// static server
	mux.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics"))))

	mux.HandleFunc("/control/push", s.handlePush)
	mux.HandleFunc("/control/pull", s.handlePull)
	mux.HandleFunc("/control/get", s.handleGet)
	mux.HandleFunc("/control/reset", s.handleReset)
	mux.HandleFunc("/control/delete", s.handleDelete)
	mux.HandleFunc("/stat/livestat", s.getLiveStatics)
	http.Serve(l, JWTMiddleware(mux))
	return nil
}

type stream struct {
	Key             string `json:"key"`
	URL             string `json:"url"`
	StreamID        uint32 `json:"stream_id"`
	VideoTotalBytes uint64 `json:"video_total_bytes"`
	VideoSpeed      uint64 `json:"video_speed"`
	AudioTotalBytes uint64 `json:"audio_total_bytes"`
	AudioSpeed      uint64 `json:"audio_speed"`
}

type streams struct {
	Publishers []stream `json:"publishers"`
	Players    []stream `json:"players"`
}

// getLiveStatics get the static of this live
func (s *Server) getLiveStatics(w http.ResponseWriter, req *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}

	defer res.SendJSON()

	rtmpStream := s.handler.(*rtmp.Streams)
	if rtmpStream == nil {
		res.Status = 500
		res.Data = "Get rtmp stream information error"
		return
	}

	msgs := new(streams)
	for item := range rtmpStream.GetStreams().IterBuffered() {
		if s, ok := item.Val.(*rtmp.Stream); ok {
			if s.Reader() != nil {
				switch s.Reader().(type) {
				case *rtmp.VirReader:
					v := s.Reader().(*rtmp.VirReader)
					msg := stream{
						Key:             item.Key,
						URL:             v.Info().URL,
						StreamID:        v.ReadBWInfo.StreamID,
						VideoTotalBytes: v.ReadBWInfo.VideoDatainBytes,
						VideoSpeed:      v.ReadBWInfo.VideoSpeedInBytesperMS,
						AudioTotalBytes: v.ReadBWInfo.AudioDatainBytes,
						AudioSpeed:      v.ReadBWInfo.AudioSpeedInBytesperMS,
					}
					msgs.Publishers = append(msgs.Publishers, msg)
				}
			}
		}
	}

	for item := range rtmpStream.GetStreams().IterBuffered() {
		ws := item.Val.(*rtmp.Stream).Ws()
		for s := range ws.IterBuffered() {
			if pw, ok := s.Val.(*rtmp.PackWriterCloser); ok {
				if pw.Writer() != nil {
					switch pw.Writer().(type) {
					case *rtmp.VirWriter:
						v := pw.Writer().(*rtmp.VirWriter)
						msg := stream{
							Key:             item.Key,
							URL:             v.Info().URL,
							StreamID:        v.WriteBWInfo.StreamID,
							VideoTotalBytes: v.WriteBWInfo.VideoDatainBytes,
							VideoSpeed:      v.WriteBWInfo.VideoSpeedInBytesperMS,
							AudioTotalBytes: v.WriteBWInfo.AudioDatainBytes,
							AudioSpeed:      v.WriteBWInfo.AudioSpeedInBytesperMS,
						}
						msgs.Players = append(msgs.Players, msg)
					}
				}
			}
		}
	}

	resp, _ := json.Marshal(msgs)
	res.Data = resp
}

// handlePull pull a rtmp stream to a application
// url schema like this:
//  http://127.0.0.1:8090/control/pull?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456
func (s *Server) handlePull(w http.ResponseWriter, req *http.Request) {
	var retString string
	var err error

	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}

	defer res.SendJSON()

	if req.ParseForm() != nil {
		res.Status = 400
		res.Data = "url: /control/pull?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456"
		return
	}

	oper := req.Form.Get("oper")
	app := req.Form.Get("app")
	name := req.Form.Get("name")
	url := req.Form.Get("url")

	log.Debugf("control pull: oper=%v, app=%v, name=%v, url=%v", oper, app, name, url)
	if (len(app) <= 0) || (len(name) <= 0) || (len(url) <= 0) {
		res.Status = 400
		res.Data = "control push parameter error, please check them."
		return
	}

	remoteurl := "rtmp://127.0.0.1" + s.rtmpAddr + "/" + app + "/" + name
	localurl := url

	keyString := "pull:" + app + "/" + name
	if oper == "stop" {
		pullRtmprelay, found := s.session[keyString]

		if !found {
			retString = fmt.Sprintf("session key[%s] not exist, please check it again.", keyString)
			res.Status = 400
			res.Data = retString
			return
		}
		log.Debugf("rtmprelay stop push %s from %s", remoteurl, localurl)
		pullRtmprelay.Stop()

		delete(s.session, keyString)
		retString = fmt.Sprintf("<h1>push url stop %s ok</h1></br>", url)
		res.Status = 400
		res.Data = retString
		log.Debugf("pull stop return %s", retString)
	} else {
		pullRtmprelay := rtmprelay.NewRtmpRelay(&localurl, &remoteurl)
		log.Debugf("rtmprelay start push %s from %s", remoteurl, localurl)
		err = pullRtmprelay.Start()
		if err != nil {
			retString = fmt.Sprintf("push error=%v", err)
		} else {
			s.session[keyString] = pullRtmprelay
			retString = fmt.Sprintf("<h1>push url start %s ok</h1></br>", url)
		}
		res.Status = 400
		res.Data = retString
		log.Debugf("pull start return %s", retString)
	}
}

// handlePush push a stream
// the URL schema like this:
//  http://127.0.0.1:8090/control/push?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456
func (s *Server) handlePush(w http.ResponseWriter, req *http.Request) {
	var retString string
	var err error

	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}

	defer res.SendJSON()

	if req.ParseForm() != nil {
		res.Data = "url: /control/push?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456"
		return
	}

	oper := req.Form.Get("oper")
	app := req.Form.Get("app")
	name := req.Form.Get("name")
	url := req.Form.Get("url")

	log.Debugf("control push: oper=%v, app=%v, name=%v, url=%v", oper, app, name, url)
	if (len(app) <= 0) || (len(name) <= 0) || (len(url) <= 0) {
		res.Data = "control push parameter error, please check them."
		return
	}

	localurl := "rtmp://127.0.0.1" + s.rtmpAddr + "/" + app + "/" + name
	remoteurl := url

	keyString := "push:" + app + "/" + name
	if oper == "stop" {
		pushRtmprelay, found := s.session[keyString]
		if !found {
			retString = fmt.Sprintf("<h1>session key[%s] not exist, please check it again.</h1>", keyString)
			res.Data = retString
			return
		}
		log.Debugf("rtmprelay stop push %s from %s", remoteurl, localurl)
		pushRtmprelay.Stop()

		delete(s.session, keyString)
		retString = fmt.Sprintf("<h1>push url stop %s ok</h1></br>", url)
		res.Data = retString
		log.Debugf("push stop return %s", retString)
	} else {
		pushRtmprelay := rtmprelay.NewRtmpRelay(&localurl, &remoteurl)
		log.Debugf("rtmprelay start push %s from %s", remoteurl, localurl)
		err = pushRtmprelay.Start()
		if err != nil {
			retString = fmt.Sprintf("push error=%v", err)
		} else {
			retString = fmt.Sprintf("<h1>push url start %s ok</h1></br>", url)
			s.session[keyString] = pushRtmprelay
		}

		res.Data = retString
		log.Debugf("push start return %s", retString)
	}
}

// handleReset reset a room
// the URL schema like:
//  http://127.0.0.1:8090/control/reset?room=ROOM_NAME
func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}
	defer res.SendJSON()

	if err := r.ParseForm(); err != nil {
		res.Status = 400
		res.Data = "url: /control/reset?room=<ROOM_NAME>"
		return
	}
	room := r.Form.Get("room")

	if len(room) == 0 {
		res.Status = 400
		res.Data = "url: /control/reset?room=<ROOM_NAME>"
		return
	}

	msg, err := configure.RoomKeys.SetKey(room)

	if err != nil {
		msg = err.Error()
		res.Status = 400
	}

	res.Data = msg
}

// handleGet get a key by channel room
// the url schema like:
//  http://127.0.0.1:8090/control/get?room=ROOM_NAME
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}
	defer res.SendJSON()

	if err := r.ParseForm(); err != nil {
		res.Status = 400
		res.Data = "url: /control/get?room=<ROOM_NAME>"
		return
	}

	room := r.Form.Get("room")

	if len(room) == 0 {
		res.Status = 400
		res.Data = "url: /control/get?room=<ROOM_NAME>"
		return
	}

	msg, err := configure.RoomKeys.GetKey(room)
	if err != nil {
		msg = err.Error()
		res.Status = 400
	}
	res.Data = msg
}

// handleDelete delete the room
// this url like this:
//   http://127.0.0.1:8090/control/delete?room=ROOM_NAME
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}
	defer res.SendJSON()

	if err := r.ParseForm(); err != nil {
		res.Status = 400
		res.Data = "url: /control/delete?room=<ROOM_NAME>"
		return
	}

	room := r.Form.Get("room")

	if len(room) == 0 {
		res.Status = 400
		res.Data = "url: /control/delete?room=<ROOM_NAME>"
		return
	}

	if configure.RoomKeys.DeleteChannel(room) {
		res.Data = "Ok"
		return
	}
	res.Status = 404
	res.Data = "room not found"
}
