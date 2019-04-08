package snowflake

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var sf *Snowflake

func init() {
	sf = NewSnowflake(Settings{})
}

func TestSonwflake(t *testing.T) {
	id, err := sf.NextID()
	assert.Nil(t, err)
	t.Log(id)

	p := Decompose(id)

	t.Log(p)

	time.Sleep(time.Millisecond * 10)

	id, err = sf.NextID()
	assert.Nil(t, err)
	t.Log(id)

	p = Decompose(id)

	t.Log(p)
}

func BenchmarkSonwflake(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := sf.NextID(); err != nil {
				return
			}
		}
	})
}
