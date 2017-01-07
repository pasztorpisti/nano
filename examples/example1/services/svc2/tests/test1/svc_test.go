package test1

import (
	"os"
	"testing"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc2"
	"github.com/pasztorpisti/nano/examples/example1/config/test"
	svc_svc1 "github.com/pasztorpisti/nano/examples/example1/services/svc1"
	svc_svc2 "github.com/pasztorpisti/nano/examples/example1/services/svc2"
	svc_svc3 "github.com/pasztorpisti/nano/examples/example1/services/svc3"
	svc_svc4 "github.com/pasztorpisti/nano/examples/example1/services/svc4"
)

var svc2Client nano.Client

func TestMain(m *testing.M) {
	test.Init()

	cs := nano.NewTestClientSet(
		svc_svc1.New(),
		svc_svc2.New(),
		svc_svc3.New(),
		svc_svc4.New(),
	)
	svc2Client = cs.LookupClient("svc2")

	os.Exit(m.Run())
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
