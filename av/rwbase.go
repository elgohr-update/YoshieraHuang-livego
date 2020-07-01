package av

import (
	"sync"
	"time"
)

type RWBase struct {
	lock               sync.Mutex
	timeout            time.Duration
	PreTime            time.Time
	BaseTimestamp      uint32
	LastVideoTimestamp uint32
	LastAudioTimestamp uint32
}

func NewRWBase(duration time.Duration) RWBase {
	return RWBase{
		timeout: duration,
		PreTime: time.Now(),
	}
}

func (rw *RWBase) BaseTimeStamp() uint32 {
	return rw.BaseTimestamp
}

func (rw *RWBase) CalcBaseTimestamp() {
	if rw.LastAudioTimestamp > rw.LastVideoTimestamp {
		rw.BaseTimestamp = rw.LastAudioTimestamp
	} else {
		rw.BaseTimestamp = rw.LastVideoTimestamp
	}
}

func (rw *RWBase) RecTimeStamp(timestamp, typeID uint32) {
	if typeID == TAG_VIDEO {
		rw.LastVideoTimestamp = timestamp
	} else if typeID == TAG_AUDIO {
		rw.LastAudioTimestamp = timestamp
	}
}

func (rw *RWBase) SetPreTime() {
	rw.lock.Lock()
	rw.PreTime = time.Now()
	rw.lock.Unlock()
}

func (rw *RWBase) Alive() bool {
	rw.lock.Lock()
	b := !(time.Now().Sub(rw.PreTime) >= rw.timeout)
	rw.lock.Unlock()
	return b
}
