package circular

import (
	"sync"
)

type Cache struct {
	idx      int32
	size     int32
	internal map[int32]interface{}
	mutex    *sync.RWMutex
}

func New(size int32) *Cache {
	return &Cache{
		idx:      0,
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

	return nil
}

func (l *Cache) Range() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	data := make([]interface{}, l.size)
	for idx, val := range l.internal {
		data[idx] = val
	}

	return data
}

func (l *Cache) RangeWithASC() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	idx := 0
	data := make([]interface{}, l.size)

	for i := l.idx - 1; i < l.size; i++ {
		data[idx] = l.internal[i]
		idx++
	}

	for i := int32(0); i < l.idx-1; i++ {
		data[idx] = l.internal[i]
		idx++
	}

	return data
}
