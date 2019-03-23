package mixins

import "testing"

func TestGet2(t *testing.T) {
	val := Get("path", "/etc/tenured")
	t.Log(val)
}
