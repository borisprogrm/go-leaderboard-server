package logger

import (
	"os"
	"reflect"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

type LogParams map[string]any

var logger = &Logger{}

func GetLogger() *Logger {
	return logger
}

func (l *Logger) Initialize(IsDebug bool) {
	if IsDebug {
		l.logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339, NoColor: true}).
			With().
			Timestamp().
			Caller().
			Logger().
			Level(zerolog.DebugLevel)
	} else {
		l.logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Caller().
			Logger().
			Level(zerolog.InfoLevel)
	}
}

func (l *Logger) processParams(e *zerolog.Event, params ...LogParams) {
	for _, param := range params {
		for k, v := range param {
			if k == "error" {
				if err, ok := v.(error); ok {
					e.Err(err)
					continue
				}
			}
			switch reflect.ValueOf(v).Kind() {
			case reflect.String:
				e.Str(k, v.(string))
			case reflect.Int:
				e.Int(k, v.(int))
			case reflect.Float64:
				e.Float64(k, v.(float64))
			default:
				panic("unknown param value type")
			}
		}
	}
}

func (l *Logger) Debug(msg string, params ...LogParams) {
	e := l.logger.Debug()
	l.processParams(e, params...)
	e.CallerSkipFrame(1).Msg(msg)
}

func (l *Logger) Info(msg string, params ...LogParams) {
	e := l.logger.Info()
	l.processParams(e, params...)
	e.CallerSkipFrame(1).Msg(msg)
}

func (l *Logger) Warn(msg string, params ...LogParams) {
	e := l.logger.Warn()
	l.processParams(e, params...)
	e.CallerSkipFrame(1).Msg(msg)
}

func (l *Logger) Error(msg string, params ...LogParams) {
	e := l.logger.Error()
	l.processParams(e, params...)
	e.CallerSkipFrame(1).Msg(msg)
}

func (l *Logger) Fatal(msg string, params ...LogParams) {
	e := l.logger.Fatal()
	l.processParams(e, params...)
	e.CallerSkipFrame(1).Msg(msg)
}
