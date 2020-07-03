package av

import (
	"sync"
	"time"
)

// RWBaser is the base for reader and writer
type RWBaser interface {
	Aliver

	BaseTimestamp() uint32
	CalcBaseTimestamp()
	RecTimestamp(timestamp, typeID uint32)
	SetPreTime()
}

// RWBase is the base for reader and writer
type RWBase struct {
	lock    sync.Mutex
	timeout time.Duration

	preTime time.Time
	// baseTimestamp is the base timestamp
	baseTimestamp uint32
	// lastVideoTimestamp is the last video timestamp
	lastVideoTimestamp uint32
	// lastAudioTimestamp is the last audio timestamp
	lastAudioTimestamp uint32
}

// NewRWBase returns a RWBaser
func NewRWBase(duration time.Duration) RWBaser {
	return &RWBase{
		timeout: duration,
		preTime: time.Now(),
	}
}

// BaseTimestamp return the base timestamp
func (rw *RWBase) BaseTimestamp() uint32 {
	return rw.baseTimestamp
}

// CalcBaseTimestamp calculates the base timestamp
func (rw *RWBase) CalcBaseTimestamp() {
	if rw.lastAudioTimestamp > rw.lastVideoTimestamp {
		rw.baseTimestamp = rw.lastAudioTimestamp
	} else {
		rw.baseTimestamp = rw.lastVideoTimestamp
	}
}

// RecTimestamp record the timestamp according to type
func (rw *RWBase) RecTimestamp(timestamp, typeID uint32) {
	if typeID == TagVideo {
		rw.lastVideoTimestamp = timestamp
	} else if typeID == TagAudio {
		rw.lastAudioTimestamp = timestamp
	}
}

// SetPreTime set the pre time
func (rw *RWBase) SetPreTime() {
	rw.lock.Lock()
	rw.preTime = time.Now()
	rw.lock.Unlock()
}

// Alive returns if this is alive
func (rw *RWBase) Alive() bool {
	rw.lock.Lock()
	b := !(time.Now().Sub(rw.preTime) >= rw.timeout)
	rw.lock.Unlock()
	return b
}
