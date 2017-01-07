/*
DO NOT EDIT!
This file has been generated from JSON by gen_http_transport_config.
*/
package svc2

import (
	"reflect"

	"github.com/pasztorpisti/nano/addons/transport/http/config"
)

var HTTPTransportConfig = &config.ServiceConfig{
	ServiceName: "svc2",
	Endpoints: []*config.EndpointConfig{
		{
			Method:        "POST",
			Path:          "/",
			HasReqContent: true,
			ReqType:       reflect.TypeOf((*Req)(nil)).Elem(),
			RespType:      reflect.TypeOf((*Resp)(nil)).Elem(),
		},
		{
			Method:        "GET",
			Path:          "/",
			HasReqContent: false,
			ReqType:       reflect.TypeOf((*GetReq)(nil)).Elem(),
			RespType:      reflect.TypeOf((*GetResp)(nil)).Elem(),
		},
	},
}
