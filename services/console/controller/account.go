package ctl

import (
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
	}
}
