package protocol

import "github.com/ihaiker/tenured-go-server/commons/remoting"

type TenuredCommandProcesser interface {
	OnCommand(channel remoting.RemotingChannel, command *TenuredCommand)
}

type tenuredCommandRunner struct {
	process     TenuredCommandProcesser
	poolSize    int
	runningSize uint32
	channel     chan struct{}
}

func (this *tenuredCommandRunner) onCommand(channel remoting.RemotingChannel, command *TenuredCommand) {

}
