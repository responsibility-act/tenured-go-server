package ctl

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/services"
	"github.com/kataras/iris/context"
	"time"
)

func init() {
	user := app.Party("/user")
	{
		user.Post("/add", tenantAuth(addUser))
		user.Get("/token/{id}", tenantAuth(requestToken))
	}
}

func addUser(app *api.App, ctx context.Context) {
	user := new(api.User)
	err := ctx.ReadJSON(user)
	if err != nil {
		writeJson(ctx, services.ErrInvalidJson)
		return
	}
	if user.TenantUserId == "" {
		writeJson(ctx, services.ErrInvalidUserId)
		return
	}

	if user.CloudId, err = id(); err != nil {
		writeJson(ctx, err)
		return
	}
	user.AccountId = app.AccountId
	user.AppId = app.Id
	user.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	user.Type = api.UserTypeNormal
	if err := UserService.AddUser(user); err != nil {
		writeJson(ctx, err)
		return
	}

}

//获取登录TOKEN
func requestToken(app *api.App, ctx context.Context) {
	userId := ctx.Params().Get("id")
	user, err := UserService.GetByTenantUserId(app.AccountId, app.Id, userId)
	if err != nil {
		writeJson(ctx, err)
		return
	}

	rt := new(api.TokenRequest)
	rt.AccountId = app.AccountId
	rt.AppId = app.Id
	rt.CloudId = user.CloudId

	if rp, err := UserService.RequestLoginToken(rt); err != nil {
		writeJson(ctx, err)
	} else {
		writeJson(ctx, rp)
	}
}
