package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/sirupsen/logrus"
)

type TenuredCommandProcesser func(channel remoting.RemotingChannel, command *TenuredCommand)

type tenuredCommandRunner struct {
	process         TenuredCommandProcesser
	executorService executors.ExecutorService
}

func (this *tenuredCommandRunner) onCommand(channel remoting.RemotingChannel, command *TenuredCommand) {
	if this.process == nil {
		logrus.Warnf("can't found command(%d) process", command.code)
		return
	}

	if this.executorService != nil {
		if err := this.executorService.Execute(func() {
			this.process(channel, command)
		}); err != nil {
			logrus.Errorf("command is error: %v", err)
		}
	} else {
		this.process(channel, command)
	}
}
