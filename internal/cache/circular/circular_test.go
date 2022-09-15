package circular

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSet(t *testing.T) {
	cache := New(2)
	assert.Equal(t, int32(0), cache.idx)

	cache.Create("1")
	assert.Equal(t, int32(1), cache.idx)

	data := cache.Range()
	assert.Equal(t, "1", data[0])

	cache.Create("2")
	assert.Equal(t, int32(2), cache.idx)

	data = cache.Range()
	assert.Equal(t, "1", data[0])
	assert.Equal(t, "2", data[1])

	cache.Create("3")
	assert.Equal(t, int32(1), cache.idx)

	data = cache.Range()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "2", data[1])

	cache.Create("4")
	assert.Equal(t, int32(2), cache.idx)

	data = cache.Range()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "4", data[1])
}

func TestSorted(t *testing.T) {
	cache := New(3)
	cache.Create(1)
	cache.Create(2)
	cache.Create(3)
	cache.Create(4)
	fmt.Println(cache.Sorted()...)
}
