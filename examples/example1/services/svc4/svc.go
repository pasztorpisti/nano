package svc4

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc4"
)

func New() nano.Service {
	return &svc{}
}

type svc struct{}

func (*svc) Name() string {
	return "svc4"
}

func (*svc) Handle(c *nano.Ctx, req interface{}) (interface{}, error) {
	switch r := req.(type) {
	case *svc4.Req:
		return &svc4.Resp{
			Value: "svc4_" + r.Param,
		}, nil
	default:
		return nil, nano.BadReqTypeError
	}
}
