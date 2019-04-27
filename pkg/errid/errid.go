package errid

import (
	"fmt"
	"net/http"
	"path"
	"runtime"
	"strings"
)

type Error struct {
	Msg      string
	ErrCause error
	ErrURL   string
	ErrTag   string
	ErrCode  int
	IsTemp   bool
	Location string
}

func New(msg string) *Error {
	// location := ""
	// _, file, no, ok := runtime.Caller(1)
	// if ok {
	// 	location = fmt.Sprintf("%s/%s:%d", path.Base(path.Dir(file)), path.Base(file), no)
	// }
	return &Error{Msg: msg, Location: getCaller(), ErrCode: defaultErrCode, ErrTag: "unknown"}
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

func (e *Error) Tag(s string) *Error {
	e.ErrTag = s
	return e
}

func (e *Error) Error() string {
	//tmpl := "error; msg=%s; loc=%s; temp=%t; code=%d; url=%s; cause=%v"
	tmpl := `{"error":{"msg":"%s", "loc":"%s", "temp":"%t", "code":%d, "url":"%s", "tag":"%s", "cause":"%v"}}`
	return fmt.Sprintf(tmpl, e.Msg, e.Location, e.IsTemp, e.ErrCode, e.ErrURL, e.ErrTag, e.ErrCause)
}

func (e *Error) GetCode() int {
	return e.ErrCode
}

func (e *Error) IsTemporary() bool {
	return e.IsTemp
}

func (e *Error) GetTag() string {
	return e.ErrTag
}

func getCaller() string {
	var sb strings.Builder
	for skip := 2; skip < 5; skip++ {
		_, file, no, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s/%s:%d", path.Base(path.Dir(file)), path.Base(file), no))
	}
	return sb.String()
}

func (e *Error) SetTag(tag string) error {
	e.ErrTag = tag
	return e
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

/////

type errtag interface {
	GetTag() string
	SetTag(tag string) error
}

func GetTag(err error) string {
	if err == nil {
		return ""
	}
	if ec, ok := err.(errtag); ok {
		return ec.GetTag()
	}
	return "unknown" // unknown error
}

func SetTag(err error, tag string) error {
	if err == nil {
		return err
	}
	if ec, ok := err.(errtag); ok {
		return ec.SetTag(tag)
	}
	return err
}
