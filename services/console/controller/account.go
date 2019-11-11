package ctl

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/kataras/iris/v12/context"
)

func id() (uint64, *protocol.TenuredError) {
	if idbody, err := clusterIdService.Get(); err != nil {
		return 0, err
	} else {
		return commons.ToUInt64(idbody), nil
	}
}

func applyAccount(ctx context.Context) {
	account := new(api.Account)
	err := ctx.ReadJSON(account)
	if err != nil {
		writeJson(ctx, err)
		return
	}
	if account.Email == "" && account.Mobile == "" {
		writeJson(ctx, protocol.NewError("AccountIsNull", "账户邮箱或者手机必填填写一项！"))
		return
	}
	account.Id, err = id()
	if err != nil {
		writeJson(ctx, err)
		return
	}
	err = accountService.Apply(account)
	if err != nil {
		writeJson(ctx, err)
		return
	}
	writeJson(ctx, nil)
}

func mobileAccount(ctx context.Context) {
	mobile := ctx.Params().Get("mobile")
	if account, err := accountService.GetByMobile(mobile); err != nil {
		writeJson(ctx, err)
	} else {
		account.Password = ""
		writeJson(ctx, account)
	}
}

func applyApp(ctx context.Context) {
	app := new(api.App)
	if err := ctx.ReadJSON(app); err != nil {
		writeJson(ctx, err)
		return
	}
	account, err := accountService.Get(app.AccountId)
	if err != nil {
		writeJson(ctx, err)
		return
	}
	if app.Name == "" {
		writeJson(ctx, protocol.NewError("AppNameIsNull", "应用名称不能为空"))
		return
	}
	logger.Infof("账户 %s 申请App: %s", account.Name, app.String())

	app.Id, err = id()
	if err != nil {
		writeJson(ctx, err)
		return
	}
	if err := accountService.ApplyApp(app); err != nil {
		writeJson(ctx, err)
	} else {
		writeJson(ctx, nil)
	}
}

func init() {
	accountServer := app.Party("/account")
	{
		accountServer.Post("/apply", applyAccount)
		accountServer.Get("/mobile/{mobile}", mobileAccount)
	}
	appServer := app.Party("/app")
	{
		appServer.Post("/apply", applyApp)
	}
}
