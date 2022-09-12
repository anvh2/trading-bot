package circular

import (
	"sync"
)

type Cache struct {
	idx      int32
	len      int32
	size     int32
	internal map[int32]interface{}
	mutex    *sync.RWMutex
}

func New(size int32) *Cache {
	return &Cache{
		idx:      0,
		len:      0,
		size:     size,
		internal: make(map[int32]interface{}, size),
		mutex:    &sync.RWMutex{},
	}
}

func (l *Cache) Set(data interface{}) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.idx >= l.size {
		l.idx -= l.size
	}

	l.internal[l.idx] = data
	l.idx++

	if l.len < l.size {
		l.len++
	}

	return nil
}

func (l *Cache) Range() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	data := make([]interface{}, l.len)
	for i := int32(0); i < l.len; i++ {
		data[i] = l.internal[i]
	}

	return data
}
