package errid

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
)

type Error struct {
	Msg      string
	ErrCause error
	ErrURL   string
	ErrCode  int
	IsTemp   bool
	Location string
}

func New(msg string) *Error {
	location := ""
	_, file, no, ok := runtime.Caller(1)
	if ok {
		location = fmt.Sprintf("%s/%s:%d", path.Base(path.Dir(file)), path.Base(file), no)
	}
	return &Error{Msg: msg, Location: location, ErrCode: defaultErrCode}
}

func (e *Error) URL(s string) *Error {
	e.ErrURL = s
	return e
}

func (e *Error) Code(code int) *Error {
	e.ErrCode = code
	return e
}

func (e *Error) Temp(isTemp bool) *Error {
	e.IsTemp = isTemp
	return e
}

func (e *Error) Cause(err error) *Error {
	e.ErrCause = err
	return e
}

func (e *Error) Error() string {
	//tmpl := "error; msg=%s; loc=%s; temp=%t; code=%d; url=%s; cause=%v"
	tmpl := `{"error":{"msg":"%s", "loc":"%s", "temp":"%t", "code":%d, "url":"%s", "cause":"%v"}}`
	return fmt.Sprintf(tmpl, e.Msg, e.Location, e.IsTemp, e.ErrCode, e.ErrURL, e.ErrCause)
}

func (e *Error) GetCode() int {
	return e.ErrCode
}

func (e *Error) IsTemporary() bool {
	return e.IsTemp
}

/////

const defaultErrCode int = http.StatusInternalServerError

type errcode interface {
	GetCode() int
}

func GetCode(err error) int {
	if err == nil {
		return 0
	}
	if ec, ok := err.(errcode); ok {
		return ec.GetCode()
	}
	return defaultErrCode // unknown error
}

/////

type temporary interface {
	IsTemporary() bool
}

func IsTemporary(err error) bool {
	if err == nil {
		return true
	}

	if ec, ok := err.(temporary); ok {
		return ec.IsTemporary()
	}
	return false // assume not temporary
}
