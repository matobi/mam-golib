package logger

import (
	"os"

	"github.com/go-kit/kit/log"
)

type Logger interface {
	Err(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Dbg(msg string, args ...interface{})
}

func Init(serviceName, profile string) Logger {
	logger := log.With(
		log.NewJSONLogger(log.NewSyncWriter(os.Stdout)),
		"time", log.DefaultTimestampUTC,
		"caller", log.Caller(5), // log.DefaultCaller,
		"name", serviceName,
		"logEnv", profile)
	w := &lw{
		logErr:  createLoggerFunc(logger, "error"),
		logInfo: createLoggerFunc(logger, "info"),
		logDbg:  createLoggerFunc(logger, "debug"),
	}
	return w
}

func createLoggerFunc(kitLogger log.Logger, level string) logfunc {
	return func(msg string, args ...interface{}) error {
		if len(args) == 0 {
			return kitLogger.Log("level", level, "msg", msg)
		}
		if _, ok := args[0].(error); ok {
			return kitLogger.Log(append(args[1:], "level", level, "msg", msg, "err", args[0])...)
		}
		return kitLogger.Log(append(args, "level", level, "msg", msg)...)
	}
}

type logfunc func(msg string, args ...interface{}) error

type lw struct {
	logErr  logfunc
	logInfo logfunc
	logDbg  logfunc
}

func (w *lw) Err(msg string, args ...interface{}) {
	w.logErr(msg, args...)
}

func (w *lw) Info(msg string, args ...interface{}) {
	w.logInfo(msg, args...)
}

func (w *lw) Dbg(msg string, args ...interface{}) {
	w.logDbg(msg, args...)
}
