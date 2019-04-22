package ctl

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/api"
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
	}

	userServer := app.Party("/user")
	{
		userServer.Post("/add", func(ctx context.Context) {
			user := &api.User{}
			if err := ctx.ReadJSON(user); err != nil {
				_, _ = ctx.WriteString(fmt.Sprintf("%v", err))
				return
			}
			sid, err := clusterIdService.Get()
			if err != nil {
				_, _ = ctx.WriteString(fmt.Sprintf("%v", err))
				return
			}
			user.ClusterId = commons.ToUInt64(sid)

			err = userService.AddUser(user)
			if commons.IsNil(err) {
				_, _ = ctx.WriteString("OK")
			} else {
				_, _ = ctx.WriteString(fmt.Sprintf("%v", err))
			}
		})
		userServer.Get("/{appId}/{tenantId}", func(ctx context.Context) {
			appId := ctx.Params().GetUint64Default("appId", 0)
			tenantId := ctx.Params().Get("tenantId")

			user, err := userService.GetByTenantUserId(1, appId, tenantId)
			if commons.IsNil(err) {
				_, _ = ctx.JSON(user)
			} else {
				_, _ = ctx.WriteString(fmt.Sprintf("%v", err))
			}
		})
	}
}
