package ctl

import (
	"github.com/kataras/iris/context"
	"strconv"
)

func tenantAuth(ctx context.Context) {
	accountId, _ := strconv.ParseUint(ctx.GetHeader("tenured_account_id"), 10, 64)
	appId, _ := strconv.ParseUint(ctx.GetHeader("tenured_app_id"), 10, 64)
	if app, err := accountService.GetApp(accountId, appId); err != nil {
		logger.Info("账户认证失败：", accountId, " err:", err)
		writeJson(ctx, errInvoildAccount)
	} else {
		assertKey := ctx.GetHeader("tenured_ak")
		sign := ctx.GetHeader("tenured_sign")
		logger.Debug("app=", app)
		logger.Debug(accountId, appId, assertKey, sign)
		ctx.Next()
	}
}

func aa(ctx context.Context) (uint64, uint64) {
	accountId, _ := strconv.ParseUint(ctx.GetHeader("tenured_account_id"), 10, 64)
	appId, _ := strconv.ParseUint(ctx.GetHeader("tenured_app_id"), 10, 64)
	return accountId, appId
}
