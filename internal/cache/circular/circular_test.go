package circular

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSet(t *testing.T) {
	cache := New(2)
	assert.Equal(t, int32(0), cache.idx)

	cache.Set("1")
	assert.Equal(t, int32(1), cache.idx)

	data := cache.Get()
	assert.Equal(t, "1", data[0])

	cache.Set("2")
	assert.Equal(t, int32(2), cache.idx)

	data = cache.Get()
	assert.Equal(t, "1", data[0])
	assert.Equal(t, "2", data[1])

	cache.Set("3")
	assert.Equal(t, int32(1), cache.idx)

	data = cache.Get()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "2", data[1])

	cache.Set("4")
	assert.Equal(t, int32(2), cache.idx)

	data = cache.Get()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "4", data[1])
}
