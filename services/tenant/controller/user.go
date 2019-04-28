package ctl

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/kataras/iris/context"
)

func init() {
	user := app.Party("/user", tenantAuth)
	{
		user.Post("/add", addUser)
		user.Post("/token", requestToken)
	}
}

func addUser(ctx context.Context) {

}

//获取登录TOKEN
func requestToken(ctx context.Context) {
	rt := new(api.TokenRequest)
	if err := ctx.ReadJSON(rt); err != nil {
		writeJson(ctx, err)
	} else {
		accountId, appId := aa(ctx)
		rt.AccountId = accountId
		rt.AppId = appId
		if rp, err := userService.RequestLoginToken(rt); err != nil {
			writeJson(ctx, err)
		} else {
			writeJson(ctx, rp)
		}
	}
}
