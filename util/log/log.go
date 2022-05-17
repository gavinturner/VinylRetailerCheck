package log

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var useTextLogs bool

// init runs on startup, and sets configuration on the default logger.
func init() {
	textLogsEnv := os.Getenv("TEXT_LOGS")
	if len(textLogsEnv) > 0 {
		useTextLogs = true
	}

	if useTextLogs {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.JSONFormatter{
			FieldMap: log.FieldMap{
				log.FieldKeyLevel: "severity",
			},
		})
	}

	// sometimes I just want to see my temporary fmt.Printf and not all the usual mess
	desiredLogLevel := os.Getenv("LOG_LEVEL")
	if lev, err := log.ParseLevel(desiredLogLevel); err != nil {
		// default unknowns and empty string to what we've been using so far in prod
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(lev)
	}
	log.SetOutput(os.Stdout)
}

type Logger struct {
	*log.Entry
}

// New creates a new logger
func New() *Logger {
	logger := log.New()
	if useTextLogs {
		logger.Formatter = &log.TextFormatter{}
	} else {
		logger.Formatter = &log.JSONFormatter{
			FieldMap: log.FieldMap{
				log.FieldKeyLevel: "severity",
			},
		}
	}

	logger.Level = log.GetLevel()
	logger.Out = os.Stdout

	return &Logger{log.NewEntry(logger)}
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Error(nil, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(nil, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Fatal(nil, format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal(nil, format, args...)
}

func Panicf(format string, args ...interface{}) {
	Panic(nil, format, args...)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Panic(nil, format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Print(args ...interface{}) {
	log.Print(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func makeErrObj(err error, format string, args ...interface{}) error {
	if err == nil {
		return fmt.Errorf(format, args...)
	} else {
		return errors.Wrapf(err, format, args...)
	}
}

func Error(err error, format string, args ...interface{}) {
	e := makeErrObj(err, format, args...)
	log.Error(e.Error())
}

func (l *Logger) Error(err error, format string, args ...interface{}) {
	e := makeErrObj(err, format, args...)
	l.Entry.Error(e.Error())
}

func Fatal(err error, format string, args ...interface{}) {
	e := makeErrObj(err, format, args...)
	log.Fatal(e.Error())
}

func (l *Logger) Fatal(err error, format string, args ...interface{}) {
	e := makeErrObj(err, format, args...)
	l.Entry.Fatal(e.Error())
}

func Panic(err error, format string, args ...interface{}) {
	e := makeErrObj(err, format, args...)
	log.Panic(e.Error())
}

func (l *Logger) Panic(err error, format string, args ...interface{}) {
	e := makeErrObj(err, format, args...)
	l.Entry.Panic(e.Error())
}

func Println(args ...interface{}) {
	log.Println(args...)
}
