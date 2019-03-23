package signal

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// InitSignal register signals handler.
func Signal(reload func()) {
	logrus.Infof("%s pid:%d", filepath.Base(os.Args[0]), os.Getpid())
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		logrus.Infof("signal : %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
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
