package dao

import (
	"github.com/ihaiker/tenured-go-server/api"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var dataDir = os.TempDir() + "/tenured"
var accountService = NewAccountServer(dataDir)

func TestAccountServer_Apply(t *testing.T) {
	t.Log(dataDir)

	err := accountService.Start()
	assert.Nil(t, err)

	account := &api.Account{}
	account.Id = 123123
	account.Name = "123123"

	err = accountService.Apply(account)
	assert.Nil(t, err)

	ac, err := accountService.Get(123123)
	assert.Nil(t, err)
	t.Log(ac)
}
