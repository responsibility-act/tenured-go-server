package protocol

import (
	"github.com/ihaiker/tenured-go-server/commons"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemotingCommand(t *testing.T) {
	request := NewRequest(1)
	assert.False(t, request.IsACK())
	assert.False(t, request.IsOneway())
	assert.Equal(t, request.Id, uint32(1))

	response := NewACK(request.Id)
	assert.True(t, response.IsACK())
	assert.True(t, response.IsSuccess())

	response.Error("1001", "test error")

	assert.False(t, response.IsSuccess())

	t.Log(response.GetError())

	response.RemotingError(commons.NewRemotingError("1002", "test error2"))

	re := response.GetError()

	assert.Equal(t, re.Code, "1002")
	t.Log(re)

	req2 := NewRequest(2)
	assert.Equal(t, req2.Id, uint32(2))
}

func TestRemotingCommand_MapHeader(t *testing.T) {
	rc := NewRequest(1)

	jsonData := map[string]string{}
	jsonData["name"] = "tenured"

	err := rc.SetHeader(&jsonData)
	assert.Nil(t, err)

	jsonData2 := map[string]string{}
	err = rc.GetHeader(&jsonData2)
	assert.Nil(t, err)
	assert.Equal(t, jsonData2["name"], "tenured")
}

type AuthInfoHeader struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func TestRemotingCommand_ObjHeader(t *testing.T) {
	rc := NewRequest(1)

	auth := &AuthInfoHeader{User: "test", Password: "pawd"}
	err := rc.SetHeader(auth)
	assert.Nil(t, err)

	authget := &AuthInfoHeader{}
	err = rc.GetHeader(authget)
	assert.Nil(t, err)
	assert.Equal(t, authget.User, auth.User)

	assert.Equal(t, auth, authget)
}
