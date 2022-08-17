package ginrest

import (
	"fmt"
)

// 错误
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

// 通过code和msg产生一个错误
func New(code int, msg string) error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

// 未知错误
func NewUnknownException(err error) error {
	return New(ErrCodeUnknownException, err.Error())
}

// 转Error
func ToError(err error) *Error {
	e, ok := err.(*Error)
	if !ok {
		panic("bad err type")
	}
	return e
}

// 获取错误的错误码
func Code(err error) int {
	return ToError(err).Code
}

// 获取trpc错误的消息
func Msg(err error) string {
	return ToError(err).Msg
}

// 错误是否保存对应错误码
func Equal(err error, code int) bool {
	return Code(err) == code
}
