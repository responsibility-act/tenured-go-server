package ctl

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/ihaiker/tenured-go-server/protocol"
	"github.com/kataras/iris/context"
	"strconv"
)

func tenantAuth(fn func(app *api.App, ctx context.Context)) context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				writeJson(ctx, protocol.ConvertError(commons.Catch(err)))
			}
		}()
		accountId, _ := strconv.ParseUint(ctx.GetHeader("tenured_account_id"), 10, 64)
		appId, _ := strconv.ParseUint(ctx.GetHeader("tenured_app_id"), 10, 64)
		if app, err := AccountService.GetApp(accountId, appId); err != nil {
			logger.Info("账户认证失败：", accountId, " err:", err)
			writeJson(ctx, errInvalidAccount)
		} else {
			assertKey := ctx.GetHeader("tenured_ak")
			sign := ctx.GetHeader("tenured_sign")
			logger.Debug("app=", app)
			logger.Debug(accountId, appId, assertKey, sign)
			fn(app, ctx)
		}
	}
}
