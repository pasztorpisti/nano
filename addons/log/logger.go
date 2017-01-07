package log

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pasztorpisti/nano"
)

type Map map[string]interface{}

type Logger interface {
	Info(c *nano.Ctx, msg string)
	Infom(c *nano.Ctx, msg string, m Map)
	Infof(c *nano.Ctx, format string, a ...interface{})
	Err(c *nano.Ctx, err error, msg string)
	Errm(c *nano.Ctx, err error, msg string, m Map)
	Errf(c *nano.Ctx, err error, format string, a ...interface{})
}

var Default Logger = NewLogger(NewWriterStream(os.Stderr))

func Info(c *nano.Ctx, msg string) {
	Default.Info(c, msg)
}

func Infom(c *nano.Ctx, msg string, m Map) {
	Default.Infom(c, msg, m)
}

func Infof(c *nano.Ctx, format string, a ...interface{}) {
	Default.Infof(c, format, a...)
}

func Err(c *nano.Ctx, err error, msg string) {
	Default.Err(c, err, msg)
}

func Errm(c *nano.Ctx, err error, msg string, m Map) {
	Default.Errm(c, err, msg, m)
}

func Errf(c *nano.Ctx, err error, format string, a ...interface{}) {
	Default.Errf(c, err, format, a...)
}

func NewLogger(s LogStream) Logger {
	return &logger{s: s}
}

type logger struct {
	mu sync.Mutex
	s  LogStream
}

const (
	typeInfo  = "INFO"
	typeError = "ERROR"
)

func (p *logger) log(c *nano.Ctx, logType, msg string) {
	t := time.Now().UTC().Format(time.RFC3339)
	clientName := "-"
	svcName := "-"
	reqID := "-"
	if c != nil {
		if c.ClientName != "" {
			clientName = c.ClientName
		}
		if c.Svc != nil {
			svcName = c.Svc.Name()
		}
		if c.ReqID != "" {
			reqID = c.ReqID
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(p.s, "%s | %s | %s | %s -> %s | %s\n",
		t, logType, reqID, clientName, svcName, msg)
	p.s.AfterLog()
}

func (p *logger) Info(c *nano.Ctx, msg string) {
	p.log(c, typeInfo, msg)
}

func (p *logger) Infom(c *nano.Ctx, msg string, m Map) {
	b := bytes.NewBuffer(nil)
	for k, v := range m {
		fmt.Fprintf(b, " %s=%#v", k, v)
	}
	p.log(c, typeInfo, msg+b.String())
}

func (p *logger) Infof(c *nano.Ctx, format string, a ...interface{}) {
	p.log(c, typeInfo, fmt.Sprintf(format, a...))
}

func (p *logger) Err(c *nano.Ctx, err error, msg string) {
	p.Errm(c, err, msg, nil)
}

func (p *logger) Errm(c *nano.Ctx, err error, msg string, m Map) {
	if err == nil && len(m) == 0 {
		p.log(c, typeError, msg)
		return
	}

	b := bytes.NewBuffer(nil)
	if err != nil {
		fmt.Fprintf(b, " error=%v", err)
	}
	for k, v := range m {
		fmt.Fprintf(b, " %s=%#v", k, v)
	}
	p.log(c, typeError, msg+b.String())
}

func (p *logger) Errf(c *nano.Ctx, err error, format string, a ...interface{}) {
	Default.Errm(c, err, fmt.Sprintf(format, a...), nil)
}
