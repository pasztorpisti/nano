/*
DO NOT EDIT!
This file has been generated from JSON by gen_http_transport_config.
*/
package svc4

import (
	"reflect"

	"github.com/pasztorpisti/nano/addons/transport/http/config"
)

var HTTPTransportConfig = &config.ServiceConfig{
	ServiceName: "svc4",
	Endpoints: []*config.EndpointConfig{
		{
			Method:        "POST",
			Path:          "/",
			HasReqContent: true,
			ReqType:       reflect.TypeOf((*Req)(nil)).Elem(),
			RespType:      reflect.TypeOf((*Resp)(nil)).Elem(),
		},
	},
}
