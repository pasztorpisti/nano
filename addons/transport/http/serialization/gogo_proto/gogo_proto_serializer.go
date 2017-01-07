package gogo_proto

import (
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"

	"github.com/gogo/protobuf/proto"
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

type reqSerializer struct{}

func (reqSerializer) SerializeRequest(ec *config.EndpointConfig,
	c *nano.Ctx, req interface{}) (h http.Header, body []byte, err error) {
	m, ok := req.(proto.Message)
	if !ok {
		err = util.Errf(nil, "expected proto.Message, got %T", req)
		return
	}
	payload, err := proto.Marshal(m)
	if err != nil {
		err = util.Err(err, "error marshaling request")
		return
	}

	request := &Request{
		Meta: &RequestMeta{
			ReqId:      c.ReqID,
			ClientName: c.ClientName,
		},
		Payload: payload,
	}
	body, err = proto.Marshal(request)
	if err != nil {
		err = util.Err(err, "error marshaling request")
		return
	}

	h = make(http.Header, 1)
	h.Set("Content-Type", "application/x-protobuf")
	return
}

type reqDeserializer struct{}

func (reqDeserializer) DeserializeRequest(ec *config.EndpointConfig,
	r *http.Request) (req interface{}, ri serialization.ReqInfo, err error) {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		err = util.ErrCode(nil, config.ErrorCodeBadRequestContentType,
			"missing request Content-Type header")
		return
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		err = util.ErrCodef(err, config.ErrorCodeBadRequestContentType,
			"error parsing request Content-Type: %q", ct)
		return
	}
	if mt != "application/x-protobuf" {
		err = util.ErrCodef(nil, config.ErrorCodeBadRequestContentType,
			"unsupported request Content-Type: %q", ct)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = util.Errf(err, "error reading request body")
		return
	}

	request := new(Request)
	err = proto.Unmarshal(body, request)
	if err != nil {
		err = util.ErrCode(err, config.ErrorCodeBadRequest,
			"error unmarshaling proto request")
		return
	}

	req = reflect.New(ec.ReqType).Interface()
	m, ok := req.(proto.Message)
	if !ok {
		err = util.Errf(nil, "request doesn't implement proto.Message: %T", req)
		return
	}
	err = proto.Unmarshal(request.Payload, m)
	if err != nil {
		err = util.ErrCodef(err, config.ErrorCodeBadRequest,
			"error unmarshaling request of type %T", req)
		return
	}

	ri.ReqID = request.Meta.ReqId
	ri.ClientName = request.Meta.ClientName
	return
}

type respSerializer struct{}

func (respSerializer) SerializeResponse(ec *config.EndpointConfig, c *nano.Ctx,
	w http.ResponseWriter, r *http.Request, resp interface{}, errResp error) error {
	if errResp != nil {
		return sendErrorResponse(w, errResp)
	}

	var body []byte
	if ec.RespType != nil {
		m, ok := resp.(proto.Message)
		if !ok {
			sendErrorResponse(w, serverError)
			return util.Errf(nil, "expected proto.Message, got %T", resp)
		}
		var err error
		body, err = proto.Marshal(m)
		if err != nil {
			sendErrorResponse(w, serverError)
			return util.Err(err, "error marshaling response")
		}
	}

	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(200)
	_, err := w.Write(body)
	if err != nil {
		return util.Err(err, "error writing marshaled response")
	}
	return nil
}

var serverError = util.ErrCode(nil, config.ErrorCodeServerError,
	"Internal Server Error")

func sendErrorResponse(w http.ResponseWriter, errResp error) error {
	code := util.GetErrCode(errResp)
	body, err := proto.Marshal(&ErrorResponse{
		Code: code,
		Msg:  errResp.Error(),
	})
	if err != nil {
		return util.Err(err, "error marshaling error response")
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
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
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		err = util.Err(nil, "missing response Content-Type header")
		return
	}
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		err = util.Errf(err, "error parsing response Content-Type: %q", ct)
		return
	}
	if mt != "application/x-protobuf" {
		err = util.Errf(nil, "unsupported response Content-Type: %q", ct)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = util.Errf(err, "error reading response body")
		return
	}

	if resp.StatusCode/100 != 2 {
		m := new(ErrorResponse)
		err = proto.Unmarshal(body, m)
		if err != nil {
			err = util.Err(err, "error unmarshaling error response")
			return
		}
		respErr = util.ErrCode(nil, m.Code, m.Msg)
		return
	}

	if ec.RespType == nil {
		if len(body) != 0 {
			err = util.Err(nil, "unexpected response body")
		}
		return
	}

	respObj = reflect.New(ec.RespType).Interface()
	m, ok := respObj.(proto.Message)
	if !ok {
		err = util.Errf(nil, "response type isn't a proto.Message: %T", respObj)
		return
	}
	err = proto.Unmarshal(body, m)
	if err != nil {
		err = util.Err(err, "error unmarshaling response")
		return
	}
	return
}
