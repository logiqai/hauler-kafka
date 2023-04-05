package log

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PackageLoggerEntry struct {
	logrus.Entry
	name       string
	Trace      func(args ...interface{})
	Tracef     func(args ...interface{})
	TraceIdStr string
}

func NewPackageLogger(level logrus.Level) *logrus.Logger {
	formatter := new(logrus.TextFormatter)
	formatter.FullTimestamp = true
	formatter.TimestampFormat = time.RFC3339Nano

	return &logrus.Logger{
		Out:       os.Stdout,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}
}

func NewPackageLoggerEntry(l *logrus.Logger, packageName string) *PackageLoggerEntry {
	p := PackageLoggerEntry{}
	p.Trace = p.traceOffFunc
	p.Tracef = p.traceOffFunc
	p.Entry = *logrus.NewEntry(l)
	p.name = packageName
	return &p
}

func NewTracer(l *logrus.Logger) *PackageLoggerEntry {
	p := PackageLoggerEntry{}
	p.Trace = p.traceOnFunc
	p.Tracef = p.traceOnFuncf
	p.Entry = *logrus.NewEntry(l)
	p.name = "Tracer"
	p.TraceIdStr = uuid.NewV4().String()
	return &p
}

// DecorateRuntimeContextForModule appends line, file and function context to the logger
func (p *PackageLoggerEntry) ContextLogger() *PackageLoggerEntry {
	if _, file, line, ok := runtime.Caller(1); ok {
		fileSplit := strings.Split(file, "/logiqai/")
		var fileN string
		if len(fileSplit) > 1 {
			fileN = fileSplit[1]
		} else {
			fileN = fileSplit[0]
		}
		return &PackageLoggerEntry{*p.WithFields(logrus.Fields{
			"File": fileN,
			"Line": line}), p.name, p.traceOffFunc, p.traceOffFunc, uuid.NewV4().String()}
	} else {
		return p
	}
}

// Throws a fatal error and exits the program
func (p *PackageLoggerEntry) FatalAndExit(err error) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		fName := runtime.FuncForPC(pc).Name()
		p.ContextLogger().WithFields(logrus.Fields{
			"File":     strings.Split(file, "/logiqai/")[1],
			"Line":     line,
			"Function": strings.Split(fName, "/logiqai/")[1],
			"Error":    err.Error(),
		}).Fatal("Fatal Error")
	} else {
		p.ContextLogger().WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Fatal("Fatal Error")
	}
}

func (p *PackageLoggerEntry) traceOnFunc(args ...interface{}) {
	if p.Logger.GetLevel() != logrus.DebugLevel {
		return
	}
	if _, file, line, ok := runtime.Caller(1); ok {
		fileSplit := strings.Split(file, "/logiqai/")
		var fileN string
		if len(fileSplit) > 1 {
			fileN = fileSplit[1]
		} else {
			fileN = fileSplit[0]
		}
		p.Debug(p.TraceIdStr, " File:", fileN, " Line:", line, " ", args)
	} else {
		p.Debug(p.TraceIdStr, args)
	}
}

func (p *PackageLoggerEntry) traceOnFuncf(args ...interface{}) {
	if p.Logger.GetLevel() != logrus.DebugLevel {
		return
	}
	if _, file, line, ok := runtime.Caller(1); ok {
		fileSplit := strings.Split(file, "/logiqai/")
		var fileN string
		if len(fileSplit) > 1 {
			fileN = fileSplit[1]
		} else {
			fileN = fileSplit[0]
		}

		template := fmt.Sprintf("%s File:%s Line:%d "+strings.Replace(args[0].(string), "%", "%%", -1), p.TraceIdStr, fileN, line)
		templateArgs := args[1:]
		p.Debugf(template, templateArgs...)
	} else {
		p.Debugf("%s "+args[0].(string), p.TraceIdStr, args[1:])
	}
}

func (p *PackageLoggerEntry) traceOffFunc(args ...interface{}) {
	return
}

func (p *PackageLoggerEntry) TraceOn() {
	p.Trace = p.traceOnFunc
}

func (p *PackageLoggerEntry) TraceOff() {
	p.Trace = p.traceOffFunc
}
