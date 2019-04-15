package leveldb

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/stretchr/testify/assert"
	"testing"
)

func searchService() *SearchServer {
	if searchService, err := NewSearchServer("/data/tenured"); err != nil {
		panic(err)
	} else if err := searchService.Start(); err != nil {
		panic(err)
	} else {
		return searchService
	}
}

func TestSearchServer_Put(t *testing.T) {
	as := searchService()

	_ = as.Remove("name")

	err := as.Put("name", []byte("value"))
	assert.Nil(t, err)

	_, err = as.Get("name1")
	assert.Equal(t, err, api.ErrSearchNotExists)

	value, err := as.Get("name")
	assert.Nil(t, err)
	assert.Equal(t, string(value), "value")
}
