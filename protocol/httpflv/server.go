package httpflv

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/rtmp"

	log "github.com/sirupsen/logrus"
)

// Server is the http flv server
type Server struct {
	handler av.Handler
}

type stream struct {
	Key string `json:"key"`
	ID  string `json:"id"`
}

type streams struct {
	Publishers []stream `json:"publishers"`
	Players    []stream `json:"players"`
}

// NewServer returns a server
func NewServer(h av.Handler) *Server {
	return &Server{
		handler: h,
	}
}

// Serve serves http requests
func (server *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleConn)
	mux.HandleFunc("/streams", server.getStream)
	http.Serve(l, mux)
	return nil
}

// getStreams get the information of publishers and players
func (server *Server) getStreams(w http.ResponseWriter, r *http.Request) *streams {
	rtmpStream := server.handler.(*rtmp.Streams)
	if rtmpStream == nil {
		return nil
	}
	msgs := new(streams)
	for item := range rtmpStream.GetStreams().IterBuffered() {
		if s, ok := item.Val.(*rtmp.Stream); ok {
			if s.Reader() != nil {
				msg := stream{item.Key, s.Reader().Info().UID}
				msgs.Publishers = append(msgs.Publishers, msg)
			}
		}
	}

	for item := range rtmpStream.GetStreams().IterBuffered() {
		ws := item.Val.(*rtmp.Stream).Ws()
		for s := range ws.IterBuffered() {
			if pw, ok := s.Val.(*rtmp.PackWriterCloser); ok {
				if pw.Writer() != nil {
					msg := stream{item.Key, pw.Writer().Info().UID}
					msgs.Players = append(msgs.Players, msg)
				}
			}
		}
	}

	return msgs
}

// getStream handles getStream request
func (server *Server) getStream(w http.ResponseWriter, r *http.Request) {
	msgs := server.getStreams(w, r)
	if msgs == nil {
		return
	}
	resp, _ := json.Marshal(msgs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// handleConn handles connections
func (server *Server) handleConn(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("http flv handleConn panic: ", r)
		}
	}()

	url := r.URL.String()
	u := r.URL.Path
	if pos := strings.LastIndex(u, "."); pos < 0 || u[pos:] != ".flv" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	path := strings.TrimSuffix(strings.TrimLeft(u, "/"), ".flv")
	paths := strings.SplitN(path, "/", 2)
	log.Debug("url:", u, "path:", path, "paths:", paths)

	if len(paths) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// 判断视屏流是否发布,如果没有发布,直接返回404
	msgs := server.getStreams(w, r)
	if msgs == nil || len(msgs.Publishers) == 0 {
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	include := false
	for _, item := range msgs.Publishers {
		if item.Key == path {
			include = true
			break
		}
	}
	if include == false {
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	writer := NewWriter(paths[0], paths[1], url, w)

	server.handler.HandleWriter(writer)
	writer.Wait()
}
