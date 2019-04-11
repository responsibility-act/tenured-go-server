//generator by tenured command defined.
package api

import (
	protocol "github.com/ihaiker/tenured-go-server/commons/protocol"
)

//RequestCode
var (
	ClusterIdServiceGet = uint16(1001)
)

//获取分布式ID
type ClusterIdService interface {
	Get() (uint64, *protocol.TenuredError)
}
