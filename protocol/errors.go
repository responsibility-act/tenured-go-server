package protocol

import "fmt"

type TenuredError struct {
	code    string
	message string
}

func (this *TenuredError) Code() string {
	return this.code
}

func (this *TenuredError) Message() string {
	return this.message
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
		code: "1000", message: "Not allowed access, not certified",
	}
}

func ErrorNoModule() *TenuredError {
	return &TenuredError{
		code: "0000", message: "Can't found module",
	}
}

func ErrorDB(err error) *TenuredError {
	return &TenuredError{code: "0001", message: err.Error()}
}

func ErrorRouter() *TenuredError {
	return &TenuredError{code: "0002", message: "No valid route"}
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
