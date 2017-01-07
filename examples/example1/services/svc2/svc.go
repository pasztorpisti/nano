package svc2

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/util"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc1"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc2"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc3"
)

func New() nano.Service {
	return &svc{}
}

type svc struct {
	svc1 nano.Client
	svc3 nano.Client
}

func (*svc) Name() string {
	return "svc2"
}

func (p *svc) Init(cs nano.ClientSet) error {
	p.svc1 = cs.LookupClient("svc1")
	p.svc3 = cs.LookupClient("svc3")
	return nil
}

func (p *svc) Handle(c *nano.Ctx, req interface{}) (interface{}, error) {
	switch r := req.(type) {
	case *svc2.Req:
		return p.handleReq(c, r)
	case *svc2.GetReq:
		return p.handleGetReq(c)
	default:
		return nil, nano.BadReqTypeError
	}
}

func (p *svc) handleReq(c *nano.Ctx, r *svc2.Req) (interface{}, error) {
	resp, err := p.svc1.Request(c, &svc1.Req{Param: r.Param})
	if err != nil {
		return nil, util.Err(err, "svc1 failure")
	}
	return &svc2.Resp{
		Value: "svc2_" + resp.(*svc1.Resp).Value,
	}, nil
}

func (p *svc) handleGetReq(c *nano.Ctx) (interface{}, error) {
	resp, err := p.svc3.Request(c, &svc3.Req{Param: "getparam"})
	if err != nil {
		return nil, util.Err(err, "svc3 failure")
	}
	return &svc2.GetResp{
		Value: "svc2_" + resp.(*svc3.Resp).Value,
	}, nil
}
