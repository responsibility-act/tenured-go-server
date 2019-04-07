package protocol

import "fmt"

type TenuredError struct {
	code    string
	message string
}

func (this *TenuredError) Code() string {
	return this.code
}

func (this *TenuredError) Is(code string) bool {
	return this.code == code
}

func (this *TenuredError) Error() string {
	return fmt.Sprintf("[%s]%s", this.code, this.message)
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
		code: "1000", message: "not found auth info.",
	}
}

func ErrorInvalidAuth() *TenuredError {
	return &TenuredError{
		code: "1001", message: "invalid auth",
	}
}

func ErrorNoModule() *TenuredError {
	return &TenuredError{
		code: "1002", message: "Can't found module",
	}
}

func ErrorInvalidHeader(err error) *TenuredError {
	return &TenuredError{code: "1003", message: err.Error()}
}

func ErrorDB(err error) *TenuredError {
	return &TenuredError{code: "1004", message: err.Error()}
}

func ErrorRouter() *TenuredError {
	return &TenuredError{code: "1005", message: "No valid route"}
}

func NewError(code, message string) *TenuredError {
	return &TenuredError{code: code, message: message}
}

func ErrorHandler(err error) *TenuredError {
	if err == nil {
		return nil
	}
	return &TenuredError{code: "9999", message: err.Error()}
}
