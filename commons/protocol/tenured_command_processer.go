package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
)

type TenuredCommandProcesser interface {
	OnCommand(channel remoting.RemotingChannel, command *TenuredCommand)
}

type tenuredCommandRunner struct {
	process         TenuredCommandProcesser
	executorService executors.ExecutorService
}

func (this *tenuredCommandRunner) onCommand(channel remoting.RemotingChannel, command *TenuredCommand) {
	if this.process == nil {
		return
	}
	if this.executorService != nil {
		_ = this.executorService.Execute(func() {
			this.process.OnCommand(channel, command)
		})
	}
}
