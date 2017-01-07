package svc1

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc1"
)

func New() nano.Service {
	return &svc{}
}

type svc struct{}

func (*svc) Name() string {
	return "svc1"
}

func (*svc) Handle(c *nano.Ctx, req interface{}) (interface{}, error) {
	switch r := req.(type) {
	case *svc1.Req:
		return &svc1.Resp{
			Value: "svc1_" + r.Param,
		}, nil
	default:
		return nil, nano.BadReqTypeError
	}
}
