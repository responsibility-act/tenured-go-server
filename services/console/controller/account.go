package ctl

import "github.com/kataras/iris/context"

func applyAccount(ctx context.Context) {

}

func init() {
	account := app.Party("/account")
	{
		account.Post("/apply", applyAccount)
	}
}
