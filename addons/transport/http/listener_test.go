package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
	json_ser "github.com/pasztorpisti/nano/addons/transport/http/serialization/json"
	"github.com/pasztorpisti/nano/addons/util"
)

const (
	listenSVCName         = "test_svc"
	listenTestReqID       = "TestReqID"
	listenTestClientName  = "test_client"
	listenJSONContentType = "application/json; charset=utf-8"
)

type ListenReq struct {
	S string
	I int
}

type ListenResp struct {
	B0 bool
	B1 bool
}

type ListenGetReq struct {
}

type ListenGetResp struct {
	S string
}

type ListenGetDirReq struct {
}

var listenCFG = &config.ServiceConfig{
	ServiceName: listenSVCName,
	Endpoints: []*config.EndpointConfig{
		{
			Method:        "POST",
			Path:          "/",
			HasReqContent: true,
			ReqType:       reflect.TypeOf((*ListenReq)(nil)).Elem(),
			RespType:      reflect.TypeOf((*ListenResp)(nil)).Elem(),
		},
		{
			Method:        "GET",
			Path:          "/",
			HasReqContent: false,
			ReqType:       reflect.TypeOf((*ListenGetReq)(nil)).Elem(),
			RespType:      reflect.TypeOf((*ListenGetResp)(nil)).Elem(),
		},
		{
			Method:        "GET",
			Path:          "/dir",
			HasReqContent: false,
			ReqType:       reflect.TypeOf((*ListenGetDirReq)(nil)).Elem(),
			RespType:      nil,
		},
	},
}

func newListener(prefixURLPath bool, h util.HandlerFunc) http.Handler {
	l := NewListener(&ListenerOptions{
		Serializer:    json_ser.ServerSideSerializer,
		PrefixURLPath: prefixURLPath,
	}, listenCFG)

	svc := util.NewService(listenSVCName, h)
	ss := nano.NewServiceSet(svc)

	err := l.Init(ss)
	if err != nil {
		panic(err)
	}

	return l.(*listener).router
}

func TestListen_Req(t *testing.T) {
	called := false
	var rc *nano.Ctx
	var ro interface{}
	h := newListener(true, func(c *nano.Ctx, req interface{}) (interface{}, error) {
		called = true
		rc = c
		ro = req
		return &ListenResp{
			B0: false,
			B1: true,
		}, nil
	})

	reqBody := `{"S":"str","I":42}`
	req := httptest.NewRequest("POST", "/"+listenSVCName+"/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", listenJSONContentType)
	req.Header.Set(json_ser.HeaderReqID, listenTestReqID)
	req.Header.Set(json_ser.HeaderClientName, listenTestClientName)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Errorf("resp.Code == %v, want %v", resp.Code, 200)
	}
	if v := resp.Header().Get("Content-Type"); v != listenJSONContentType {
		t.Errorf("Content-Type header == %q, want %q", v, listenJSONContentType)
	}

	respObj := new(ListenResp)
	err := json.Unmarshal(resp.Body.Bytes(), respObj)
	if err != nil {
		t.Errorf("error unmarshaling response content :: %v", err)
		t.FailNow()
	}
	if respObj.B0 || !respObj.B1 {
		t.Errorf("unexpected response content: %#v", respObj)
	}

	if !called {
		t.Error("request handler wasn't called")
		t.FailNow()
	}
	if rc == nil {
		t.Error("request handler received nil *nano.Ctx")
		t.FailNow()
	}
	if rc.ReqID != listenTestReqID {
		t.Errorf("rc.ReqID == %q, want %q", rc.ReqID, listenTestReqID)
	}
	if rc.ClientName != listenTestClientName {
		t.Errorf("rc.ClientName == %q, want %q", rc.ClientName, listenTestClientName)
	}
	reqObj, ok := ro.(*ListenReq)
	if ok {
		if reqObj.S != "str" {
			t.Errorf("reqObj.S == %q, want %q", reqObj.S, "str")
		}
		if reqObj.I != 42 {
			t.Errorf("reqObj.I == %v, want %v", reqObj.I, 42)
		}
	} else {
		t.Errorf("handle received object of type %T, want %T", ro, reqObj)
	}
}

func TestListen_GetReq(t *testing.T) {
	called := false
	var rc *nano.Ctx
	var ro interface{}
	h := newListener(true, func(c *nano.Ctx, req interface{}) (interface{}, error) {
		called = true
		rc = c
		ro = req
		return &ListenGetResp{
			S: "str",
		}, nil
	})

	req := httptest.NewRequest("GET", "/"+listenSVCName+"/", nil)
	req.Header.Set(json_ser.HeaderReqID, listenTestReqID)
	req.Header.Set(json_ser.HeaderClientName, listenTestClientName)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Errorf("resp.Code == %v, want %v", resp.Code, 200)
	}
	if v := resp.Header().Get("Content-Type"); v != listenJSONContentType {
		t.Errorf("Content-Type header == %q, want %q", v, listenJSONContentType)
	}

	respObj := new(ListenGetResp)
	err := json.Unmarshal(resp.Body.Bytes(), respObj)
	if err != nil {
		t.Errorf("error unmarshaling response content :: %v", err)
		t.FailNow()
	}
	if respObj.S != "str" {
		t.Errorf("unexpected response content: %#v", respObj)
	}

	if !called {
		t.Error("request handler wasn't called")
		t.FailNow()
	}
	if rc == nil {
		t.Error("request handler received nil *nano.Ctx")
		t.FailNow()
	}
	if rc.ReqID != listenTestReqID {
		t.Errorf("rc.ReqID == %q, want %q", rc.ReqID, listenTestReqID)
	}
	if rc.ClientName != listenTestClientName {
		t.Errorf("rc.ClientName == %q, want %q", rc.ClientName, listenTestClientName)
	}
	reqObj, ok := ro.(*ListenGetReq)
	if !ok {
		t.Errorf("handle received object of type %T, want %T", ro, reqObj)
	}
}

func TestListen_GetDirReq(t *testing.T) {
	called := false
	var rc *nano.Ctx
	var ro interface{}
	h := newListener(true, func(c *nano.Ctx, req interface{}) (interface{}, error) {
		called = true
		rc = c
		ro = req
		return nil, nil
	})

	req := httptest.NewRequest("GET", "/"+listenSVCName+"/dir", nil)
	req.Header.Set(json_ser.HeaderReqID, listenTestReqID)
	req.Header.Set(json_ser.HeaderClientName, listenTestClientName)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Errorf("resp.Code == %v, want %v", resp.Code, 200)
	}
	if v := resp.Header().Get("Content-Type"); v != "" {
		t.Errorf("unexpected Content-Type header: %q", v)
	}

	if resp.Body.Len() != 0 {
		t.Errorf("unexpected response body: %v", resp.Body.Bytes())
	}

	if !called {
		t.Error("request handler wasn't called")
		t.FailNow()
	}
	if rc == nil {
		t.Error("request handler received nil *nano.Ctx")
		t.FailNow()
	}
	if rc.ReqID != listenTestReqID {
		t.Errorf("rc.ReqID == %q, want %q", rc.ReqID, listenTestReqID)
	}
	if rc.ClientName != listenTestClientName {
		t.Errorf("rc.ClientName == %q, want %q", rc.ClientName, listenTestClientName)
	}
	reqObj, ok := ro.(*ListenGetDirReq)
	if !ok {
		t.Errorf("handle received object of type %T, want %T", ro, reqObj)
	}
}

func TestListen_WithoutPrefixURLPath(t *testing.T) {
	called := false
	var rc *nano.Ctx
	var ro interface{}
	h := newListener(false, func(c *nano.Ctx, req interface{}) (interface{}, error) {
		called = true
		rc = c
		ro = req
		return nil, nil
	})

	req := httptest.NewRequest("GET", "/dir", nil)
	req.Header.Set(json_ser.HeaderReqID, listenTestReqID)
	req.Header.Set(json_ser.HeaderClientName, listenTestClientName)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != 200 {
		t.Errorf("resp.Code == %v, want %v", resp.Code, 200)
	}
	if v := resp.Header().Get("Content-Type"); v != "" {
		t.Errorf("unexpected Content-Type header: %q", v)
	}

	if resp.Body.Len() != 0 {
		t.Errorf("unexpected response body: %v", resp.Body.Bytes())
	}

	if !called {
		t.Error("request handler wasn't called")
		t.FailNow()
	}
	if rc == nil {
		t.Error("request handler received nil *nano.Ctx")
		t.FailNow()
	}
	if rc.ReqID != listenTestReqID {
		t.Errorf("rc.ReqID == %q, want %q", rc.ReqID, listenTestReqID)
	}
	if rc.ClientName != listenTestClientName {
		t.Errorf("rc.ClientName == %q, want %q", rc.ClientName, listenTestClientName)
	}
	reqObj, ok := ro.(*ListenGetDirReq)
	if !ok {
		t.Errorf("handle received object of type %T, want %T", ro, reqObj)
	}
}

func TestListen_MethodNotAllowed(t *testing.T) {
	h := newListener(true, func(c *nano.Ctx, req interface{}) (interface{}, error) {
		t.Error("handler was called")
		return nil, nil
	})

	reqBody := `{"S":"str","I":42}`
	req := httptest.NewRequest("PATCH", "/"+listenSVCName+"/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", listenJSONContentType)
	req.Header.Set(json_ser.HeaderReqID, listenTestReqID)
	req.Header.Set(json_ser.HeaderClientName, listenTestClientName)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusMethodNotAllowed {
		t.Errorf("resp.Code == %v, want %v", resp.Code, http.StatusMethodNotAllowed)
		t.FailNow()
	}
}

func testListen_ErrorResponse(t *testing.T, errCode string, expectedStatus int) {
	const errMsg = "test error"
	called := false
	h := newListener(true, func(c *nano.Ctx, req interface{}) (interface{}, error) {
		called = true
		return nil, util.ErrCode(nil, errCode, errMsg)
	})

	req := httptest.NewRequest("GET", "/"+listenSVCName+"/", nil)
	req.Header.Set(json_ser.HeaderReqID, listenTestReqID)
	req.Header.Set(json_ser.HeaderClientName, listenTestClientName)

	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if !called {
		t.Error("handler wasn't called")
	}

	if resp.Code != expectedStatus {
		t.Errorf("resp.Code == %v, want %v", resp.Code, expectedStatus)
	}
	if v := resp.Header().Get("Content-Type"); v != listenJSONContentType {
		t.Errorf("Content-Type header == %q, want %q", v, listenJSONContentType)
	}

	respObj := new(json_ser.ErrorResponse)
	err := json.Unmarshal(resp.Body.Bytes(), respObj)
	if err != nil {
		t.Errorf("error unmarshaling error response content :: %v", err)
		t.FailNow()
	}
	if respObj.Code != errCode || respObj.Msg != errMsg {
		t.Errorf("unexpected response content: %#v", respObj)
	}
}

func TestListen_NotFoundErrorResponse(t *testing.T) {
	testListen_ErrorResponse(t, config.ErrorCodeNotFound, 404)
}

func TestListen_ClientErrorResponse(t *testing.T) {
	testListen_ErrorResponse(t, "C-MYERROR", 400)
}

func TestListen_ServerErrorResponse(t *testing.T) {
	testListen_ErrorResponse(t, "S-MYERROR", 500)
}
