/*
DO NOT EDIT!
This file has been generated from JSON by gen_http_transport_config.
*/
package svc3

import (
	"reflect"

	"github.com/pasztorpisti/nano/addons/transport/http/config"
)

var HTTPTransportConfig = &config.ServiceConfig{
	ServiceName: "svc3",
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
