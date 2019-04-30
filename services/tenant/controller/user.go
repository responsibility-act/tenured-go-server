package ctl

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/kataras/iris/context"
)

func init() {
	user := app.Party("/user")
	{
		user.Post("/add", tenantAuth(addUser))
		user.Post("/token", tenantAuth(requestToken))
	}
}

func addUser(app *api.App, ctx context.Context) {

}

//获取登录TOKEN
func requestToken(app *api.App, ctx context.Context) {
	rt := new(api.TokenRequest)
	if err := ctx.ReadJSON(rt); err != nil {
		writeJson(ctx, err)
	} else {
		rt.AccountId = app.AccountId
		rt.AppId = app.Id
		if rp, err := UserService.RequestLoginToken(rt); err != nil {
			writeJson(ctx, err)
		} else {
			writeJson(ctx, rp)
		}
	}
}
