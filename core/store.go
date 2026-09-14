package core

import (
	"log"
	"time"

	"github.com/NamanG22/Redis/config"
)

var store map[string]*Obj

type Obj struct {
	Value interface{}
	ExpiresAt int64
}

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, expirationMs int64) *Obj {
	var expiresAt int64 = -1
	if expirationMs > 0 {
		expiresAt = time.Now().UnixMilli() + expirationMs
	}
	return &Obj{Value: value, ExpiresAt: expiresAt}
}

func Put(key string, obj *Obj) {
	// here I have hardcoded the max key limit, in redis this is configurable to memory limits
	if len(store) >= config.MaxKeyLimit {
		evict()
	}
	store[key] = obj
}

func Get(key string) *Obj {
	obj := store[key]
	if obj != nil {
		if obj.ExpiresAt > 0 && obj.ExpiresAt < time.Now().UnixMilli() {
			delete(store, key)
			log.Println("Key", key, "was passively deleted")
			return nil
		}	
	}
	return obj
}

func Delete(key string) bool {
	if _, ok := store[key]; !ok {
		return false
	}
	delete(store, key)
	return true
}