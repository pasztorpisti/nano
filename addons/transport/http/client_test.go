package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/discovery/static"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
	json_ser "github.com/pasztorpisti/nano/addons/transport/http/serialization/json"
	"github.com/pasztorpisti/nano/addons/util"
)

const (
	clientSVCName         = "test_svc"
	clientTestReqID       = "TestReqID"
	clientTestClientName  = "test"
	clientJSONContentType = "application/json; charset=utf-8"
)

type ClientReq struct {
	S string
	I int
}

type ClientResp struct {
	B0 bool
	B1 bool
}

type ClientGetReq struct {
}

type ClientGetResp struct {
	S string
}

type ClientGetDirReq struct {
}

var clientCFG = &config.ServiceConfig{
	ServiceName: clientSVCName,
	Endpoints: []*config.EndpointConfig{
		{
			Method:        "POST",
			Path:          "/",
			HasReqContent: true,
			ReqType:       reflect.TypeOf((*ClientReq)(nil)).Elem(),
			RespType:      reflect.TypeOf((*ClientResp)(nil)).Elem(),
		},
		{
			Method:        "GET",
			Path:          "/",
			HasReqContent: false,
			ReqType:       reflect.TypeOf((*ClientGetReq)(nil)).Elem(),
			RespType:      reflect.TypeOf((*ClientGetResp)(nil)).Elem(),
		},
		{
			Method:        "GET",
			Path:          "/dir",
			HasReqContent: false,
			ReqType:       reflect.TypeOf((*ClientGetDirReq)(nil)).Elem(),
			RespType:      nil,
		},
	},
}

func newClient(prefixURLPath bool, h http.HandlerFunc) (client nano.Service, cleanup func()) {
	server := httptest.NewServer(h)

	// transport that routes all traffic to the test server
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	opts := &ClientOptions{
		Client:        &http.Client{Transport: transport},
		Discoverer:    static.Discoverer{clientSVCName: clientSVCName + ":8000"},
		Serializer:    json_ser.ClientSideSerializer,
		PrefixURLPath: prefixURLPath,
	}

	client = NewClient(opts, clientCFG)
	cleanup = server.Close
	return
}

func newCtx(svc nano.Service) *nano.Ctx {
	return &nano.Ctx{
		ReqID:      clientTestReqID,
		Context:    context.Background(),
		Svc:        svc,
		ClientName: clientTestClientName,
	}
}

func TestClient_Req(t *testing.T) {
	called := false
	client, cleanup := newClient(true, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if path, want := r.URL.Path, "/"+clientSVCName+"/"; path != want {
			t.Errorf("req path == %q, want %q", path, want)
		}
		if v := r.Header.Get("Content-Type"); v != clientJSONContentType {
			t.Errorf("req Content-Type == %q, want %q", v, clientJSONContentType)
		}
		if v := r.Header.Get(json_ser.HeaderReqID); v != clientTestReqID {
			t.Errorf("req id == %q, want %q", v, clientTestReqID)
		}
		if v := r.Header.Get(json_ser.HeaderClientName); v != clientTestClientName {
			t.Errorf("req client name == %q, want %q", v, clientTestClientName)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading request body: %v", err)
			return
		}

		req := make(map[string]interface{})
		err = json.Unmarshal(reqBody, &req)
		if err != nil {
			t.Errorf("error unmarshaling request :: %v", err)
		} else if len(req) != 2 || req["S"] != "str" || req["I"] != 42.0 {
			t.Errorf("wrong req json: %q", string(reqBody))
		}

		respBody, err := json.Marshal(&ClientResp{
			B0: false,
			B1: true,
		})
		if err != nil {
			t.Errorf("error marshaling response :: %v", err)
			return
		}
		w.Header().Set("Content-Type", clientJSONContentType)
		w.WriteHeader(200)
		_, err = w.Write(respBody)
		if err != nil {
			t.Errorf("error writing response :: %v", err)
		}
	})
	defer cleanup()

	c := newCtx(client)
	req := &ClientReq{
		S: "str",
		I: 42,
	}
	resp, err := client.Handle(c, req)
	if err != nil {
		t.Errorf("client error :: %v", err)
		t.FailNow()
	}
	if !called {
		t.Error("handler wasn't called")
	}

	respObj, ok := resp.(*ClientResp)
	if !ok {
		t.Errorf("client returned response object of type %T, want %T", resp, respObj)
		t.FailNow()
	}

	if respObj.B0 || !respObj.B1 {
		t.Errorf("wrong response object value: %#v", respObj)
	}
}

func TestClient_GetReq(t *testing.T) {
	called := false
	client, cleanup := newClient(true, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if path, want := r.URL.Path, "/"+clientSVCName+"/"; path != want {
			t.Errorf("req path == %q, want %q", path, want)
		}
		if v := r.Header.Get("Content-Type"); v != "" {
			t.Errorf("unexpected Content-Type header: %q", v)
		}
		if v := r.Header.Get(json_ser.HeaderReqID); v != clientTestReqID {
			t.Errorf("req id == %q, want %q", v, clientTestReqID)
		}
		if v := r.Header.Get(json_ser.HeaderClientName); v != clientTestClientName {
			t.Errorf("req client name == %q, want %q", v, clientTestClientName)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading request body: %v", err)
			return
		}

		if len(reqBody) != 0 {
			t.Error("unexpected request body")
		}

		respBody, err := json.Marshal(&ClientGetResp{
			S: "str",
		})
		if err != nil {
			t.Errorf("error marshaling response :: %v", err)
			return
		}
		w.Header().Set("Content-Type", clientJSONContentType)
		w.WriteHeader(200)
		_, err = w.Write(respBody)
		if err != nil {
			t.Errorf("error writing response :: %v", err)
		}
	})
	defer cleanup()

	c := newCtx(client)
	resp, err := client.Handle(c, &ClientGetReq{})
	if err != nil {
		t.Errorf("client error :: %v", err)
		t.FailNow()
	}
	if !called {
		t.Error("handler wasn't called")
	}

	respObj, ok := resp.(*ClientGetResp)
	if !ok {
		t.Errorf("client returned response object of type %T, want %T", resp, respObj)
		t.FailNow()
	}

	if respObj.S != "str" {
		t.Errorf("wrong response object value: %#v", respObj)
	}
}

func TestClient_GetDirReq(t *testing.T) {
	called := false
	client, cleanup := newClient(true, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if path, want := r.URL.Path, "/"+clientSVCName+"/dir"; path != want {
			t.Errorf("req path == %q, want %q", path, want)
		}
		if v := r.Header.Get("Content-Type"); v != "" {
			t.Errorf("unexpected Content-Type header: %q", v)
		}
		if v := r.Header.Get(json_ser.HeaderReqID); v != clientTestReqID {
			t.Errorf("req id == %q, want %q", v, clientTestReqID)
		}
		if v := r.Header.Get(json_ser.HeaderClientName); v != clientTestClientName {
			t.Errorf("req client name == %q, want %q", v, clientTestClientName)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading request body: %v", err)
			return
		}

		if len(reqBody) != 0 {
			t.Error("unexpected request body")
		}
	})
	defer cleanup()

	c := newCtx(client)
	resp, err := client.Handle(c, &ClientGetDirReq{})
	if err != nil {
		t.Errorf("client error :: %v", err)
		t.FailNow()
	}
	if !called {
		t.Error("handler wasn't called")
	}

	if resp != nil {
		t.Errorf("resp == %v, want nil", resp)
	}
}

func testClient_ErrorResponse(t *testing.T, errCode string) {
	const errMsg = "test error"
	called := false
	client, cleanup := newClient(true, func(w http.ResponseWriter, r *http.Request) {
		called = true
		if path, want := r.URL.Path, "/"+clientSVCName+"/dir"; path != want {
			t.Errorf("req path == %q, want %q", path, want)
		}
		if v := r.Header.Get("Content-Type"); v != "" {
			t.Errorf("unexpected Content-Type header: %q", v)
		}
		if v := r.Header.Get(json_ser.HeaderReqID); v != clientTestReqID {
			t.Errorf("req id == %q, want %q", v, clientTestReqID)
		}
		if v := r.Header.Get(json_ser.HeaderClientName); v != clientTestClientName {
			t.Errorf("req client name == %q, want %q", v, clientTestClientName)
		}

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading request body: %v", err)
			return
		}

		if len(reqBody) != 0 {
			t.Error("unexpected request body")
		}

		respBody, err := json.Marshal(&json_ser.ErrorResponse{
			Code: errCode,
			Msg:  errMsg,
		})
		if err != nil {
			t.Errorf("error marshaling response :: %v", err)
			return
		}
		w.Header().Set("Content-Type", clientJSONContentType)
		w.WriteHeader(config.ErrorCodeToHTTPStatus(errCode))
		_, err = w.Write(respBody)
		if err != nil {
			t.Errorf("error writing response :: %v", err)
		}
	})
	defer cleanup()

	c := newCtx(client)
	_, err := client.Handle(c, &ClientGetDirReq{})
	if err == nil {
		t.Error("unexpected success")
		t.FailNow()
	}

	if !called {
		t.Error("handler wasn't called")
	}

	if v := util.GetErrCode(err); v != errCode {
		t.Errorf("error code == %q, want %q", v, errCode)
	}
	if err.Error() != errMsg {
		t.Errorf("error msg == %q, want %q", err.Error(), errMsg)
	}
}

func TestClient_GenericErrorResponse(t *testing.T) {
	testClient_ErrorResponse(t, "")
}

func TestClient_NotFoundErrorResponse(t *testing.T) {
	testClient_ErrorResponse(t, config.ErrorCodeNotFound)
}

func TestClient_ClientErrorResponse(t *testing.T) {
	testClient_ErrorResponse(t, "C-MYERROR")
}

func TestClient_ServerErrorResponse(t *testing.T) {
	testClient_ErrorResponse(t, "S-MYERROR")
}
