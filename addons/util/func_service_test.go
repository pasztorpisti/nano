package util

import (
	"errors"
	"testing"

	"github.com/pasztorpisti/nano"
)

var handlerError = errors.New("handler")
var initError = errors.New("init")
var initFinishedError = errors.New("initFinished")

func emptyHandler(c *nano.Ctx, req interface{}) (resp interface{}, err error) {
	return nil, handlerError
}

func emptyInit(cs nano.ClientSet) error {
	return initError
}

func emptyInitFinished() error {
	return initFinishedError
}

func TestNewService(t *testing.T) {
	svc := NewService("svc", emptyHandler)

	if _, err := svc.Handle(nil, nil); err != handlerError {
		t.Errorf("svc.Handle returned an unexpected error: %v", err)
	}

	if _, ok := svc.(nano.ServiceInit); ok {
		t.Error("svc implements nano.ServiceInit")
	}

	if _, ok := svc.(nano.ServiceInitFinished); ok {
		t.Error("svc implements nano.ServiceInitFinished")
	}
}

func TestNewServiceOpts(t *testing.T) {
	svc := NewServiceOpts(ServiceOpts{
		Name:    "svc",
		Handler: emptyHandler,
	})

	if _, err := svc.Handle(nil, nil); err != handlerError {
		t.Errorf("svc.Handle returned an unexpected error: %v", err)
	}

	if _, ok := svc.(nano.ServiceInit); ok {
		t.Error("svc implements nano.ServiceInit")
	}

	if _, ok := svc.(nano.ServiceInitFinished); ok {
		t.Error("svc implements nano.ServiceInitFinished")
	}
}

func TestNewServiceOpts_Init(t *testing.T) {
	svc := NewServiceOpts(ServiceOpts{
		Name:    "svc",
		Handler: emptyHandler,
		Init:    emptyInit,
	})

	if _, err := svc.Handle(nil, nil); err != handlerError {
		t.Errorf("svc.Handle returned an unexpected error: %v", err)
	}

	if svcInit, ok := svc.(nano.ServiceInit); ok {
		if err := svcInit.Init(nil); err != initError {
			t.Errorf("svc.Init returned an unexpected error: %v", err)
		}
	} else {
		t.Error("svc doesn't implement nano.ServiceInit")
	}

	if _, ok := svc.(nano.ServiceInitFinished); ok {
		t.Error("svc implements nano.ServiceInitFinished")
	}
}

func TestNewServiceOpts_InitFinished(t *testing.T) {
	svc := NewServiceOpts(ServiceOpts{
		Name:         "svc",
		Handler:      emptyHandler,
		InitFinished: emptyInitFinished,
	})

	if _, err := svc.Handle(nil, nil); err != handlerError {
		t.Errorf("svc.Handle returned an unexpected error: %v", err)
	}

	if _, ok := svc.(nano.ServiceInit); ok {
		t.Error("svc implements nano.ServiceInit")
	}

	if svcInitFinished, ok := svc.(nano.ServiceInitFinished); ok {
		if err := svcInitFinished.InitFinished(); err != initFinishedError {
			t.Errorf("svc.InitFinished returned an unexpected error: %v", err)
		}
	} else {
		t.Error("svc doesn't implement nano.ServiceInitFinished")
	}
}

func TestNewServiceOpts_Init_And_InitFinished(t *testing.T) {
	svc := NewServiceOpts(ServiceOpts{
		Name:         "svc",
		Handler:      emptyHandler,
		Init:         emptyInit,
		InitFinished: emptyInitFinished,
	})

	if _, err := svc.Handle(nil, nil); err != handlerError {
		t.Errorf("svc.Handle returned an unexpected error: %v", err)
	}

	if svcInit, ok := svc.(nano.ServiceInit); ok {
		if err := svcInit.Init(nil); err != initError {
			t.Errorf("svc.Init returned an unexpected error: %v", err)
		}
	} else {
		t.Error("svc doesn't implement nano.ServiceInit")
	}

	if svcInitFinished, ok := svc.(nano.ServiceInitFinished); ok {
		if err := svcInitFinished.InitFinished(); err != initFinishedError {
			t.Errorf("svc.InitFinished returned an unexpected error: %v", err)
		}
	} else {
		t.Error("svc doesn't implement nano.ServiceInitFinished")
	}
}
