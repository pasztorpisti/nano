package util

import "fmt"

type NanoError interface {
	error
	Code() string
}

var ErrMsgChainSeparator = " :: "

func ErrCode(cause error, code, msg string) NanoError {
	e := &nanoError{
		cause: cause,
		code:  code,
		msg:   msg,
	}
	if e.code == "" && cause != nil {
		e.code = GetErrCode(cause)
	}
	return e
}

func ErrCodef(cause error, code, format string, a ...interface{}) NanoError {
	return ErrCode(cause, code, fmt.Sprintf(format, a))
}

func Err(cause error, msg string) NanoError {
	return ErrCode(cause, "", msg)
}

func Errf(cause error, format string, a ...interface{}) NanoError {
	return ErrCodef(cause, "", format, a)
}

func GetErrCode(err error) string {
	if e, ok := err.(NanoError); ok {
		return e.Code()
	}
	return ""
}

type nanoError struct {
	cause error
	code  string
	msg   string
}

func (p *nanoError) Error() string {
	msg := p.msg
	if p.cause != nil {
		msg += ErrMsgChainSeparator + p.cause.Error()
	}
	return msg
}

func (p *nanoError) Code() string {
	return p.code
}
