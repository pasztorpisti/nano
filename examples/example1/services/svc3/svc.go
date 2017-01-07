package svc3

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/util"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc3"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc4"
)

func New() nano.Service {
	return &svc{}
}

type svc struct {
	svc4 nano.Client
}

func (*svc) Name() string {
	return "svc3"
}

func (p *svc) Init(cs nano.ClientSet) error {
	p.svc4 = cs.LookupClient("svc4")
	return nil
}

func (p *svc) Handle(c *nano.Ctx, req interface{}) (interface{}, error) {
	switch r := req.(type) {
	case *svc3.Req:
		return p.handleReq(c, r)
	default:
		return nil, nano.BadReqTypeError
	}
}

func (p *svc) handleReq(c *nano.Ctx, r *svc3.Req) (interface{}, error) {
	resp, err := p.svc4.Request(c, &svc4.Req{Param: r.Param})
	if err != nil {
		return nil, util.Err(err, "svc4 failure")
	}
	return &svc3.Resp{
		Value: "svc3_" + resp.(*svc4.Resp).Value,
	}, nil
}
