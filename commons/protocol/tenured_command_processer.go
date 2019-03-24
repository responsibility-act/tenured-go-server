package protocol

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/sirupsen/logrus"
	"time"
)

type TenuredCommandProcesser func(channel remoting.RemotingChannel, request *TenuredCommand)

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
			start := time.Now().Unix()
			this.process(channel, command)
			fmt.Println(time.Now().Unix() - start)
		}); err != nil {
			logrus.Errorf("command is error: %v", err)
		}
	} else {
		this.process(channel, command)
	}
}
