package nets

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestRandPort(t *testing.T) {
	port, err := RandPort("", 8080, 9090)
	assert.Nil(t, err)
	assert.Equal(t, port, 8080)

	lis, err := net.Listen("tcp", ":8080")
	if err == nil {
		port, err := RandPort("", 8080, 9090)
		assert.Nil(t, err)
		assert.Equal(t, port, 8081)
	} else {
		_ = lis.Close()
	}
}

func TestIpAndPort_Address(t *testing.T) {
	f := IpAndPort{
		Port:              6072,
		IgnoredInterfaces: []string{"docker0", "无线网络连接"},
		PreferredNetworks: []string{"192.168.203.*"},
	}
	t.Log(f.GetAddress())
	t.Log(f.GetExternal())
}
