package linker

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/commons/executors"
	"github.com/ihaiker/tenured-go-server/commons/protocol"
	"github.com/ihaiker/tenured-go-server/commons/registry"
	"github.com/ihaiker/tenured-go-server/commons/remoting"
	"github.com/sirupsen/logrus"
	"time"
)

func Server(reg registry.RegistryPlugins) (commons.Service, error) {

	config := remoting.DefaultConfig()
	config.IdleTime = 1
	server, _ := protocol.NewTenuredServer(":9998", config)
	server.AuthChecker = nil

	executorService := executors.NewFixedExecutorService(1, 100)

	server.RegisterCommandProcesser(uint16(2), func(channel remoting.RemotingChannel, command *protocol.TenuredCommand) {
		ack := protocol.NewACK(command.ID())
		body := "no body"
		if command.Body != nil {
			body += string(command.Body)
		}
		logrus.
			WithField("ip", channel.RemoteAddr()).
			WithField("code", 2).
			Infof("body: %s", body)

		ack.Body = []byte(body)

		if err := channel.Write(ack, time.Second); err != nil {
			logrus.Error("write err:", err)
		}
	}, executorService)

	server.RegisterCommandProcesser(uint16(3), func(channel remoting.RemotingChannel, command *protocol.TenuredCommand) {
		header := map[string]string{}
		if err := command.GetHeader(&header); err != nil {
			logrus.
				WithField("ip", channel.RemoteAddr()).
				WithField("code", 3).
				Info("reader header error:", err)
		}
		header["tenured"] = time.Now().Format("2006-01-02")

		ack := protocol.NewACK(command.ID())
		if err := ack.SetHeader(header); err != nil {
			logrus.
				WithField("ip", channel.RemoteAddr()).
				WithField("code", 3).
				Info("set header error:", err)
		} else {
			if err := channel.Write(ack, time.Second); err != nil {
				logrus.
					WithField("ip", channel.RemoteAddr()).
					WithField("code", 3).
					Error("write err:", err)
			}
		}

	}, executorService)

	server.RegisterCommandProcesser(uint16(4), func(channel remoting.RemotingChannel, command *protocol.TenuredCommand) {
		ack := protocol.NewACK(command.ID())
		ack.RemotingError(protocol.ErrorInvalidAuth())
		if err := channel.Write(ack, time.Second); err != nil {
			logrus.
				WithField("ip", channel.RemoteAddr()).
				WithField("code", 4).
				Error("write err:", err)
		}
	}, executorService)

	return server, nil
}
