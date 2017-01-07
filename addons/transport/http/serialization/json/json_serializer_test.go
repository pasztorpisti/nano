package json

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
	"github.com/pasztorpisti/nano/addons/util"
)

type ReqNoContent struct{}

var endpointConfigNoContent = &config.EndpointConfig{
	Method:        "GET",
	Path:          "/path",
	HasReqContent: false,
	ReqType:       reflect.TypeOf((*ReqNoContent)(nil)).Elem(),
	RespType:      nil,
}

type ReqWithContent struct {
	S string
	I int
}

type RespWithContent struct {
	B0 bool
	B1 bool
}

var endpointConfigWithContent = &config.EndpointConfig{
	Method:        "POST",
	Path:          "/path",
	HasReqContent: true,
	ReqType:       reflect.TypeOf((*ReqWithContent)(nil)).Elem(),
	RespType:      reflect.TypeOf((*RespWithContent)(nil)).Elem(),
}

const (
	testReqID       = "TestReqID"
	testClientName  = "test"
	jsonContentType = "application/json; charset=utf-8"
)

func newCtx() *nano.Ctx {
	return &nano.Ctx{
		ReqID:      testReqID,
		Context:    context.Background(),
		ClientName: testClientName,
	}
}

func TestReqSerialization_NoContent(t *testing.T) {
	c := newCtx()
	ec := endpointConfigNoContent
	h, body, err := ClientSideSerializer.ReqSerializer.SerializeRequest(
		ec, c, &ReqNoContent{})
	if err != nil {
		t.Errorf("SerializeRequest failed :: %v", err)
		t.FailNow()
	}

	if v := h.Get(HeaderReqID); v != testReqID {
		t.Errorf("header(%v) == %q, want %q", HeaderReqID, v, testReqID)
	}
	if v := h.Get(HeaderClientName); v != testClientName {
		t.Errorf("header(%v) == %q, want %q", HeaderClientName, v, testClientName)
	}
	if v := h.Get("Content-Type"); v != "" {
		t.Errorf("unexpected Content-Type header: %q", v)
	}
	if len(body) > 0 {
		t.Errorf("unexpected content: %v", body)
	}

	if t.Failed() {
		t.FailNow()
	}

	r := httptest.NewRequest(ec.Method, ec.Path, nil)
	r.Header = h
	req, ri, err := ServerSideSerializer.ReqDeserializer.DeserializeRequest(ec, r)
	if err != nil {
		t.Errorf("DeserializeRequest failed :: %v", err)
		t.FailNow()
	}
	if reflect.TypeOf(req) != reflect.PtrTo(ec.ReqType) {
		t.Errorf("deserialised req type == %v, want %v", reflect.TypeOf(req),
			reflect.PtrTo(ec.ReqType))
	}
	if ri.ReqID != testReqID {
		t.Errorf("deserialised req id == %q, want %q", ri.ReqID, testReqID)
	}
	if ri.ClientName != testClientName {
		t.Errorf("deserialised req client name == %q, want %q", ri.ClientName, testClientName)
	}
}

func TestReqSerialization_WithContent(t *testing.T) {
	c := newCtx()
	ec := endpointConfigWithContent
	inputReq := &ReqWithContent{
		S: "str",
		I: 42,
	}
	h, body, err := ClientSideSerializer.ReqSerializer.SerializeRequest(
		ec, c, inputReq)
	if err != nil {
		t.Errorf("SerializeRequest failed :: %v", err)
		t.FailNow()
	}

	if v := h.Get(HeaderReqID); v != testReqID {
		t.Errorf("header(%v) == %q, want %q", HeaderReqID, v, testReqID)
	}
	if v := h.Get(HeaderClientName); v != testClientName {
		t.Errorf("header(%v) == %q, want %q", HeaderClientName, v, testClientName)
	}
	if v := h.Get("Content-Type"); v != jsonContentType {
		t.Errorf("Content-Type header == %q, want %q", v, jsonContentType)
	}
	if len(body) == 0 {
		t.Error("no content")
	}

	if t.Failed() {
		t.FailNow()
	}

	r := httptest.NewRequest(ec.Method, ec.Path, bytes.NewReader(body))
	r.Header = h
	reqObj, ri, err := ServerSideSerializer.ReqDeserializer.DeserializeRequest(ec, r)
	if err != nil {
		t.Errorf("DeserializeRequest failed :: %v", err)
		t.FailNow()
	}
	if reflect.TypeOf(reqObj) != reflect.PtrTo(ec.ReqType) {
		t.Errorf("deserialised req type == %v, want %v", reflect.TypeOf(reqObj),
			reflect.PtrTo(ec.ReqType))
	}
	req := reqObj.(*ReqWithContent)
	if inputReq.S != req.S || inputReq.I != req.I {
		t.Errorf("desrialised req == %#v, want %#v", req, inputReq)
	}
	if ri.ReqID != testReqID {
		t.Errorf("deserialised req id == %q, want %q", ri.ReqID, testReqID)
	}
	if ri.ClientName != testClientName {
		t.Errorf("deserialised req client name == %q, want %q", ri.ClientName, testClientName)
	}
}

func TestRespSerialization_NoContent(t *testing.T) {
	c := newCtx()
	ec := endpointConfigNoContent
	w := httptest.NewRecorder()
	r := httptest.NewRequest(ec.Method, ec.Path, nil)
	err := ServerSideSerializer.SerializeResponse(ec, c, w, r, nil, nil)
	if err != nil {
		t.Errorf("SerializeResponse failed :: %v", err)
		t.FailNow()
	}

	if w.Code != 200 {
		t.Errorf("response status code %v, want %v", w.Code, 200)
	}
	if v := w.Header().Get("Content-Type"); v != "" {
		t.Errorf("unexpected Content-Type header: %q", v)
	}
	if w.Body.Len() > 0 {
		t.Errorf("unexpected content: %v", w.Body.Bytes())
	}

	if t.Failed() {
		t.FailNow()
	}

	respObj, respErr, err := ClientSideSerializer.RespDeserializer.DeserializeResponse(ec, c, w.Result())
	if err != nil {
		t.Errorf("DeserializeResponse failed :: %v", err)
		t.FailNow()
	}
	if respErr != nil {
		t.Errorf("unexpected respErr :: %v", respErr)
	}
	if respObj != nil {
		t.Errorf("unexpected respObj == %#v, want nil", respObj)
	}
}

func TestRespSerialization_WithContent(t *testing.T) {
	c := newCtx()
	ec := endpointConfigWithContent
	w := httptest.NewRecorder()
	r := httptest.NewRequest(ec.Method, ec.Path, nil)
	inputResp := &RespWithContent{
		B0: false,
		B1: true,
	}
	err := ServerSideSerializer.SerializeResponse(ec, c, w, r, inputResp, nil)
	if err != nil {
		t.Errorf("SerializeResponse failed :: %v", err)
		t.FailNow()
	}

	if w.Code != 200 {
		t.Errorf("response status code %v, want %v", w.Code, 200)
	}
	if v := w.Header().Get("Content-Type"); v != jsonContentType {
		t.Errorf("Content-Type header == %q, want %q", v, jsonContentType)
	}
	if w.Body.Len() == 0 {
		t.Error("no response content")
	}

	if t.Failed() {
		t.FailNow()
	}

	respObj, respErr, err := ClientSideSerializer.RespDeserializer.DeserializeResponse(ec, c, w.Result())
	if err != nil {
		t.Errorf("DeserializeResponse failed :: %v", err)
		t.FailNow()
	}
	if respErr != nil {
		t.Errorf("unexpected respErr :: %v", respErr)
	}
	if reflect.TypeOf(respObj) != reflect.PtrTo(ec.RespType) {
		t.Errorf("deserialised resp type == %v, want %v", reflect.TypeOf(respObj),
			reflect.PtrTo(ec.RespType))
	}
	resp := respObj.(*RespWithContent)
	if inputResp.B0 != resp.B0 || inputResp.B1 != resp.B1 {
		t.Errorf("desrialised resp == %#v, want %#v", resp, inputResp)
	}
}

func testErrorResponseSerialization(t *testing.T, e error, status int) {
	c := newCtx()
	ec := endpointConfigNoContent
	w := httptest.NewRecorder()
	r := httptest.NewRequest(ec.Method, ec.Path, nil)
	err := ServerSideSerializer.SerializeResponse(ec, c, w, r, nil, e)
	if err != nil {
		t.Errorf("SerializeResponse failed :: %v", err)
		t.FailNow()
	}

	if w.Code != status {
		t.Errorf("response status code %v, want %v", w.Code, status)
	}
	if v := w.Header().Get("Content-Type"); v != jsonContentType {
		t.Errorf("Content-Type header == %q, want %q", v, jsonContentType)
	}
	if w.Body.Len() == 0 {
		t.Error("no content")
	}

	if t.Failed() {
		t.FailNow()
	}

	respObj, respErr, err := ClientSideSerializer.RespDeserializer.DeserializeResponse(ec, c, w.Result())
	if err != nil {
		t.Errorf("DeserializeResponse failed :: %v", err)
		t.FailNow()
	}
	if respObj != nil {
		t.Errorf("unexpected respObj == %#v, want nil", respObj)
	}
	if respErr == nil {
		t.Error("respErr is nil")
		t.FailNow()
	}
	if respErr.Error() != e.Error() {
		t.Errorf("error message == %q, want %q", respErr.Error(), e.Error())
	}
	errCode := util.GetErrCode(e)
	if v := util.GetErrCode(respErr); v != errCode {
		t.Errorf("error code == %q, want %q", v, errCode)
	}
}

func TestRespSerialization_NonUtilError(t *testing.T) {
	testErrorResponseSerialization(t, errors.New("test error"), 500)
}

func TestRespSerialization_ClientUtilError(t *testing.T) {
	// Since the error code has a "C-" prefix it will be treated as a client
	// error so the http status will be 400 instead of 500.
	e := util.ErrCode(nil, "C-WHATEVER", "test error")
	testErrorResponseSerialization(t, e, 400)
}

func TestRespSerialization_ServerUtilError(t *testing.T) {
	// Error code prefixes other than "C-" should result in 500.
	e := util.ErrCode(nil, "S-WHATEVER", "test error")
	testErrorResponseSerialization(t, e, 500)
}

func TestRespSerialization_NotFoundUtilError(t *testing.T) {
	// The config.ErrorCodeNotFound is handled by the http transport
	// implementation specially and it results in http status 404.
	e := util.ErrCode(nil, config.ErrorCodeNotFound, "test error")
	testErrorResponseSerialization(t, e, 404)
}
