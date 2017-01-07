package test2

import (
	"os"
	"testing"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/util"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc1"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc2"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc3"
	"github.com/pasztorpisti/nano/examples/example1/config/test"
	svc_svc2 "github.com/pasztorpisti/nano/examples/example1/services/svc2"
)

var svc2Client nano.Client

func TestMain(m *testing.M) {
	test.Init()

	svc1Mock := util.NewService("svc1", func(c *nano.Ctx, req interface{}) (interface{}, error) {
		return &svc1.Resp{Value: "svc1_" + req.(*svc1.Req).Param}, nil
	})

	// The svc1 and svc3 mocks demonstrate two different ways of creating mocks.
	cs := nano.NewTestClientSet(
		svc1Mock,
		&svc3Mock{},
		svc_svc2.New(),
	)
	svc2Client = cs.LookupClient("svc2")

	os.Exit(m.Run())
}

type svc3Mock struct{}

func (*svc3Mock) Name() string {
	return "svc3"
}

func (p *svc3Mock) Handle(c *nano.Ctx, req interface{}) (interface{}, error) {
	switch r := req.(type) {
	case *svc3.Req:
		return &svc3.Resp{Value: "svc3_svc4_" + r.Param}, nil
	default:
		return nil, nano.BadReqTypeError
	}
}

func TestReq(t *testing.T) {
	respObj, err := svc2Client.Request(nil, &svc2.Req{Param: "testval"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}

	resp, ok := respObj.(*svc2.Resp)
	if !ok {
		t.Errorf("response type == %T, want %T", resp, &svc2.Resp{})
		t.FailNow()
	}

	want := "svc2_svc1_testval"
	if resp.Value != want {
		t.Errorf("response value == %q, want %q", resp.Value, want)
	}
}

func TestGetReq(t *testing.T) {
	respObj, err := svc2Client.Request(nil, &svc2.GetReq{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		t.FailNow()
	}

	resp, ok := respObj.(*svc2.GetResp)
	if !ok {
		t.Errorf("response type == %T, want %T", resp, &svc2.GetResp{})
		t.FailNow()
	}

	want := "svc2_svc3_svc4_getparam"
	if resp.Value != want {
		t.Errorf("response value == %q, want %q", resp.Value, want)
	}
}

func TestBadRequest(t *testing.T) {
	_, err := svc2Client.Request(nil, nil)
	if err.Error() != nano.BadReqTypeError.Error() {
		t.Errorf("err == %v, want %v", err, nano.BadReqTypeError)
	}
}
