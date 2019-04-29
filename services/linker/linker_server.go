package linker

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/ihaiker/tenured-go-server/registry/load_balance"
)

type LinkerCommandHanler struct {
	sessionManager protocol.SessionManager
}

//获取当前连接点的连接数,返回个数，使用了uint64转码了
func (this *LinkerCommandHanler) GetLinkedCount(gl *load_balance.GlobalLoading) ([]byte, *protocol.TenuredError) {
	count := this.sessionManager.Size()
	return commons.Int32(int32(count)), nil
}

func NewLinkerCommandHandler(sessionManager protocol.SessionManager) *LinkerCommandHanler {
	return &LinkerCommandHanler{sessionManager: sessionManager}
}
