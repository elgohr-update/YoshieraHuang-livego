package configure

import (
	"fmt"

	"github.com/gwuhaolin/livego/utils/uid"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// RoomKeysType is a storage for room channel name and key
type RoomKeysType struct {
	localCache *cache.Cache
}

// RoomKeys storages room channel and key
var RoomKeys = &RoomKeysType{
	localCache: cache.New(cache.NoExpiration, 0),
}

// SetKey set a random key for channel
func (r *RoomKeysType) SetKey(channel string) (key string, err error) {
	for {
		key = uid.RandStringRunes(48)
		if _, found := r.localCache.Get(key); !found {
			r.localCache.SetDefault(channel, key)
			r.localCache.SetDefault(key, channel)
			break
		}
	}
	return
}

// GetKey get a key for channel
func (r *RoomKeysType) GetKey(channel string) (newKey string, err error) {
	var key interface{}
	var found bool
	if key, found = r.localCache.Get(channel); found {
		return key.(string), nil
	}
	newKey, err = r.SetKey(channel)
	log.Debugf("[KEY] new channel [%s]: %s", channel, newKey)
	return
}

// GetChannel get channel name by key
func (r *RoomKeysType) GetChannel(key string) (channel string, err error) {
	chann, found := r.localCache.Get(key)
	if found {
		return chann.(string), nil
	}

	return "", fmt.Errorf("%s does not exists", key)
}

// DeleteChannel delete channel
func (r *RoomKeysType) DeleteChannel(channel string) bool {
	key, ok := r.localCache.Get(channel)
	if ok {
		r.localCache.Delete(channel)
		r.localCache.Delete(key.(string))
		return true
	}
	return false
}

// DeleteKey delete key
func (r *RoomKeysType) DeleteKey(key string) bool {
	channel, ok := r.localCache.Get(key)
	if ok {
		r.localCache.Delete(channel.(string))
		r.localCache.Delete(key)
		return true
	}
	return false
}
