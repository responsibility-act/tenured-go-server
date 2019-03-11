package protocol

import "fmt"

type TenuredError struct {
	Code    string
	Message string
}

func (this *TenuredError) Error() string {
	return fmt.Sprintf("[%s]%s", this.Code, this.Message)
}
