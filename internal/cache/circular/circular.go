package circular

import (
	"sync"
)

type Cache struct {
	idx      int32
	size     int32
	internal map[int32]string
	mutex    *sync.RWMutex
}

func New(size int32) *Cache {
	return &Cache{
		idx:      0,
		size:     size,
		internal: make(map[int32]string, size),
		mutex:    &sync.RWMutex{},
	}
}

func (l *Cache) Set(data string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.idx >= l.size {
		l.idx -= l.size
	}

	l.internal[l.idx] = data
	l.idx++

	return nil
}

func (l *Cache) Range() []string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	data := make([]string, l.size)
	for idx, val := range l.internal {
		data[idx] = val
	}

	return data
}

func (l *Cache) RangeWithASC() []string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	idx := 0
	data := make([]string, l.size)

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
