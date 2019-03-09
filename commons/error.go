package commons

import (
	"fmt"
)

type Error string

func (this Error) Error() string {
	return string(this)
}

type RemotingError struct {
	Code    string
	Message string
}

func (this *RemotingError) Error() string {
	return fmt.Sprintf("[%s]%s", this.Code, this.Message)
}

func NewRemotingError(code, message string) RemotingError {
	return RemotingError{Code: code, Message: message}
}
