package signal

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// InitSignal register signals handler.
func Signal(reload func()) {
	logrus.Infof("%s pid:%d", filepath.Base(os.Args[0]), os.Getpid())
	c := make(chan os.Signal, 1)
	//linux
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-c
		logrus.Infof("signal: %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			if reload != nil {
				reload()
			}
		default:
			return
		}
	}
}
