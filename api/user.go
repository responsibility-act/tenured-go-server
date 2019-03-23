package api

import "github.com/ihaiker/tenured-go-server/api/command"

type UserService interface {
	AddOrUpdateUser(user command.User) (cloudId string, err error)
}
