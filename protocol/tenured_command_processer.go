package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons"
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
			this.processCommand(channel, command)
		}); err != nil {
			logger.Errorf("command is error: %v", err)
		}
	} else {
		this.processCommand(channel, command)
	}
}

func (this *tenuredCommandRunner) processCommand(channel remoting.RemotingChannel, request *TenuredCommand) {
	commons.Try(func() {
		this.process(channel, request)
	}, func(e error) {
		logger.Errorf("process %d error: %s", request.code, e.Error())
	})
}
