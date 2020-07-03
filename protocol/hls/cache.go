package hls

import (
	"bytes"
	"container/list"
	"fmt"
	"sync"
)

const (
	maxTSCacheNum = 3
)

var (
	// ErrNoKey means no key
	ErrNoKey = fmt.Errorf("No key for cache")
)

// TSCacheItem is the ts cache item
type TSCacheItem struct {
	id   string
	num  int
	lock sync.RWMutex
	ll   *list.List
	lm   map[string]TSItem
}

// NewTSCacheItem returns a TSCacheItem
func NewTSCacheItem(id string) *TSCacheItem {
	return &TSCacheItem{
		id:  id,
		ll:  list.New(),
		num: maxTSCacheNum,
		lm:  make(map[string]TSItem),
	}
}

// ID returns the ID
func (tsCacheItem *TSCacheItem) ID() string {
	return tsCacheItem.id
}

// GenM3U8PlayList generates m3u8 playlist
// TODO: found data race, fix it
func (tsCacheItem *TSCacheItem) GenM3U8PlayList() ([]byte, error) {
	var seq int
	var getSeq bool
	var maxDuration int
	m3u8body := bytes.NewBuffer(nil)
	for e := tsCacheItem.ll.Front(); e != nil; e = e.Next() {
		key := e.Value.(string)
		v, ok := tsCacheItem.lm[key]
		if ok {
			if v.Duration > maxDuration {
				maxDuration = v.Duration
			}
			if !getSeq {
				getSeq = true
				seq = v.SeqNum
			}
			fmt.Fprintf(m3u8body, "#EXTINF:%.3f,\n%s\n", float64(v.Duration)/float64(1000), v.Name)
		}
	}
	w := bytes.NewBuffer(nil)
	fmt.Fprintf(w,
		"#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-ALLOW-CACHE:NO\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:%d\n\n",
		maxDuration/1000+1, seq)
	w.Write(m3u8body.Bytes())
	return w.Bytes(), nil
}

// SetItem set item with key
func (tsCacheItem *TSCacheItem) SetItem(key string, item TSItem) {
	if tsCacheItem.ll.Len() == tsCacheItem.num {
		e := tsCacheItem.ll.Front()
		tsCacheItem.ll.Remove(e)
		k := e.Value.(string)
		delete(tsCacheItem.lm, k)
	}
	tsCacheItem.lm[key] = item
	tsCacheItem.ll.PushBack(key)
}

// GetItem get item by key
func (tsCacheItem *TSCacheItem) GetItem(key string) (TSItem, error) {
	item, ok := tsCacheItem.lm[key]
	if !ok {
		return item, ErrNoKey
	}
	return item, nil
}
