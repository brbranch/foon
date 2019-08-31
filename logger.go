package foon

import (
	"context"
	"google.golang.org/appengine/log"
	"fmt"
)

type Logger interface {
	Trace(message string)
	Warning(message string)
}

type defaultLogger struct {
	c context.Context
}

func (d defaultLogger) Trace(message string) {
	//log.Infof(d.c, "%s", message)
}

func (d defaultLogger) Warning(message string) {
	log.Warningf(d.c, "%s", message)
}

func tracef(logger Logger, format string, args ...interface{}) {
	if logger == nil {
		return
	}
	//logger.Trace(fmt.Sprintf(format, args...))
}

func warningf(logger Logger, format string, args ...interface{}) {
	if logger == nil {
		return
	}
	logger.Warning(fmt.Sprintf(format, args...))
}
