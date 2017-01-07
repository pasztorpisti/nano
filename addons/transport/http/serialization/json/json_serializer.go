package json

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"strings"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
	"github.com/pasztorpisti/nano/addons/transport/http/serialization"
	"github.com/pasztorpisti/nano/addons/util"
)

var ClientSideSerializer = &serialization.ClientSideSerializer{
	ReqSerializer:    &reqSerializer{},
	RespDeserializer: &respDeserializer{},
}

var ServerSideSerializer = &serialization.ServerSideSerializer{
	ReqDeserializer: &reqDeserializer{},
	RespSerializer:  &respSerializer{},
}

type ErrorResponse struct {
	Code string `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

const (
	HeaderReqID      = "X-Nano-Req-Id"
	HeaderClientName = "X-Nano-Client-Name"
)

type reqSerializer struct{}

func (reqSerializer) SerializeRequest(ec *config.EndpointConfig,
	c *nano.Ctx, req interface{}) (h http.Header, body []byte, err error) {
	if ec.HasReqContent {
		body, err = json.Marshal(req)
		if err != nil {
			err = util.Err(err, "error marshaling request")
			return
		}
	}

	h = make(http.Header, 3)
	if ec.HasReqContent {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	if c.ReqID != "" {
		h.Set(HeaderReqID, c.ReqID)
	}
	if c.ClientName != "" {
		h.Set(HeaderClientName, c.ClientName)
	}
	return
}

type reqDeserializer struct{}

func (reqDeserializer) DeserializeRequest(ec *config.EndpointConfig,
	r *http.Request) (req interface{}, ri serialization.ReqInfo, err error) {
	if ec.HasReqContent {
		ct := r.Header.Get("Content-Type")
		if ct == "" {
			err = util.ErrCode(nil, config.ErrorCodeBadRequestContentType,
				"missing request Content-Type header")
			return
		}
		mt, mp, err2 := mime.ParseMediaType(ct)
		if err2 != nil {
			err = util.ErrCodef(err2, config.ErrorCodeBadRequestContentType,
				"error parsing request Content-Type: %q", ct)
			return
		}
		if mt != "application/json" {
			err = util.ErrCodef(nil, config.ErrorCodeBadRequestContentType,
				"unsupported request Content-Type: %q", ct)
			return
		}
		if cs, ok := mp["charset"]; ok && strings.ToLower(cs) != "utf-8" {
			err = util.ErrCodef(nil, config.ErrorCodeBadRequestContentType,
				"unsupported request json charset: %v", cs)
			return
		}
	}

	req = reflect.New(ec.ReqType).Interface()

	if ec.HasReqContent {
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			err = util.Err(err, "error reading request body")
			return
		}
		err = json.Unmarshal(body, req)
		if err != nil {
			err = util.ErrCodef(err, config.ErrorCodeBadRequest,
				"error unmarshaling request of type %T", req)
			return
		}
	} else if r.ContentLength > 0 {
		err = util.Err(nil, "unexpected request content")
		return
	} else if r.ContentLength < 0 {
		buf := make([]byte, 1)
		n, err2 := r.Body.Read(buf)
		if err2 != nil && err2 != io.EOF {
			err = util.Err(err2, "error reading request body")
			return
		}
		if n != 0 {
			err = util.Err(nil, "unexpected request content")
			return
		}
	}

	ri.ReqID = r.Header.Get(HeaderReqID)
	ri.ClientName = r.Header.Get(HeaderClientName)
	return
}

type respSerializer struct{}

func (respSerializer) SerializeResponse(ec *config.EndpointConfig, c *nano.Ctx,
	w http.ResponseWriter, r *http.Request, resp interface{}, errResp error) error {
	if errResp != nil {
		return sendErrorResponse(w, errResp)
	}

	if ec.RespType == nil {
		return nil
	}

	var body []byte
	body, err := json.Marshal(resp)
	if err != nil {
		sendErrorResponse(w, serverError)
		return util.Err(err, "error marshaling response")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	_, err = w.Write(body)
	if err != nil {
		return util.Err(err, "error writing marshaled response")
	}
	return nil
}

var serverError = util.ErrCode(nil, config.ErrorCodeServerError,
	"Internal Server Error")

func sendErrorResponse(w http.ResponseWriter, errResp error) error {
	code := util.GetErrCode(errResp)
	body, err := json.Marshal(&ErrorResponse{
		Code: code,
		Msg:  errResp.Error(),
	})
	if err != nil {
		return util.Err(err, "error marshaling error response")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(config.ErrorCodeToHTTPStatus(code))
	_, err = w.Write(body)
	if err != nil {
		return util.Err(err, "error writing error response")
	}
	return nil
}

type respDeserializer struct{}

func (respDeserializer) DeserializeResponse(ec *config.EndpointConfig, c *nano.Ctx,
	resp *http.Response) (respObj interface{}, respErr error, err error) {
	if ec.RespType != nil {
		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			if resp.StatusCode/100 != 2 {
				err = util.Errf(nil, "HTTP status: %v", resp.Status)
				return
			}
			err = util.Err(nil, "missing response Content-Type header")
			return
		}
		mt, mp, err2 := mime.ParseMediaType(ct)
		if err2 != nil {
			err = util.Errf(err2, "error parsing response Content-Type: %q", ct)
			return
		}
		if mt != "application/json" {
			if resp.StatusCode/100 != 2 {
				err = util.Errf(nil, "HTTP status: %v", resp.Status)
				return
			}
			err = util.Errf(nil, "unsupported response Content-Type: %q", ct)
			return
		}
		if cs, ok := mp["charset"]; ok && strings.ToLower(cs) != "utf-8" {
			err = util.Errf(nil, "unsupported response json charset: %v", cs)
			return
		}
	}

	if resp.StatusCode/100 != 2 {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			err = util.Errf(err2, "error reading response body")
			return
		}
		m := new(ErrorResponse)
		err = json.Unmarshal(body, m)
		if err != nil {
			err = util.Err(err, "error unmarshaling error response")
			return
		}
		respErr = util.ErrCode(nil, m.Code, m.Msg)
		return
	}

	if ec.RespType != nil {
		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			err = util.Errf(err2, "error reading response body")
			return
		}
		respObj = reflect.New(ec.RespType).Interface()
		err = json.Unmarshal(body, respObj)
		if err != nil {
			err = util.Err(err, "error unmarshaling response")
			return
		}
		return
	} else if resp.ContentLength > 0 {
		err = util.Err(nil, "unexpected response content")
		return
	} else if resp.ContentLength < 0 {
		buf := make([]byte, 1)
		n, err2 := resp.Body.Read(buf)
		if err2 != nil && err2 != io.EOF {
			err = util.Err(err2, "error reading response body")
			return
		}
		if n != 0 {
			err = util.Err(nil, "unexpected response content")
			return
		}
		return
	}

	return
}
