package c8tmap

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var n = New()

func TestConcurrentMap(t *testing.T) {
	s := sync.WaitGroup{}
	for j := 0; j < 10; j++ {
		s.Add(1)
		a := j * 100
		go func() {
			for i := 0; i < 100; i++ {
				n.Set(a+i, a+i)
			}
			s.Done()
		}()
	}
	s.Wait()
	assert.Equal(t, n.Count(), 1000)
}

func TestConcurrentMap_Items(t *testing.T) {
	TestConcurrentMap(t)
	it := n.IterBuffered()
	for tu := <-it; tu.Key != nil; tu = <-it {
		n.Pop(tu.Key)
	}
	assert.Equal(t, n.Count(), 0)
}
