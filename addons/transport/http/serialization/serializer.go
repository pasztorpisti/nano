package serialization

import (
	"net/http"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
)

type ReqSerializer interface {
	SerializeRequest(ec *config.EndpointConfig, c *nano.Ctx, req interface{},
	) (h http.Header, body []byte, err error)
}

type ReqInfo struct {
	ReqID      string
	ClientName string
}

type ReqDeserializer interface {
	DeserializeRequest(ec *config.EndpointConfig, r *http.Request,
	) (req interface{}, ri ReqInfo, err error)
}

type RespSerializer interface {
	// c might be nil if errResp!=nil.
	SerializeResponse(ec *config.EndpointConfig, c *nano.Ctx, w http.ResponseWriter,
		r *http.Request, resp interface{}, errResp error) error
}

type RespDeserializer interface {
	DeserializeResponse(ec *config.EndpointConfig, c *nano.Ctx, resp *http.Response,
	) (respObj interface{}, respErr error, err error)
}

type ClientSideSerializer struct {
	ReqSerializer
	RespDeserializer
}

type ServerSideSerializer struct {
	ReqDeserializer
	RespSerializer
}
