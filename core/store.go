package core

import "time"

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
	store[key] = obj
}

func Get(key string) *Obj {
	return store[key]
}