package ctl

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/kataras/iris/context"
)

func applyAccount(ctx context.Context) {
	if err := accountService.Apply(nil); err != nil {
		writeJson(ctx, err)
	} else {
		writeJson(ctx, nil)
	}
}

func init() {
	accountServer := app.Party("/account")
	{
		accountServer.Post("/apply", applyAccount)
		accountServer.Get("/id", func(ctx context.Context) {
			id, err := clusterIdService.Get()
			if err != nil {
				logger.Warn("get id error: ", err)
				writeJson(ctx, err)
			} else {
				idUint := commons.ToUInt64(id)
				logger.Debug("get id := ", idUint)
				writeJson(ctx, &struct {
					Id uint64 `json:"id"`
				}{Id: idUint})
			}
		})
	}
}
