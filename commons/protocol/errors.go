package protocol

import "fmt"

type TenuredError struct {
	Code    string
	Message string
}

func (this *TenuredError) Error() string {
	return fmt.Sprintf("[%s]%s", this.Code, this.Message)
}

func ConvertError(err error) *TenuredError {
	if terr, ok := err.(*TenuredError); ok {
		return terr
	} else {
		return ErrorHandler(err)
	}
}

func ErrorNoAuth() *TenuredError {
	return &TenuredError{
		Code: "1000", Message: "not found auth info.",
	}
}

func ErrorInvalidAuth() *TenuredError {
	return &TenuredError{
		Code: "1001", Message: "invalid auth",
	}
}

func ErrorNoModule() *TenuredError {
	return &TenuredError{
		Code: "1002", Message: "Can't found module",
	}
}

func ErrorInvalidHeader(err error) *TenuredError {
	return &TenuredError{Code: "1003", Message: err.Error()}
}

func ErrorHandler(err error) *TenuredError {
	return &TenuredError{Code: "9999", Message: err.Error()}
}
