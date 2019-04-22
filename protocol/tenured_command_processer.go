package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
)

type TenuredCommandProcesser func(channel remoting.RemotingChannel, request *TenuredCommand)

type tenuredCommandRunner struct {
	process         TenuredCommandProcesser
	executorService executors.ExecutorService
}

func (this *tenuredCommandRunner) onCommand(channel remoting.RemotingChannel, command *TenuredCommand) {
	if this.process == nil {
		logger.Warnf("can't found command(%d) process", command.code)
		return
	}

	if this.executorService != nil {
		if err := this.executorService.Execute(func() {
			this.process(channel, command)
		}); err != nil {
			logger.Errorf("command is error: %v", err)
		}
	} else {
		this.process(channel, command)
	}
}
