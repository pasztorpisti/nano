package config

import "reflect"

type ServiceConfig struct {
	ServiceName string
	Endpoints   []*EndpointConfig
}

type EndpointConfig struct {
	Method        string
	Path          string
	HasReqContent bool
	ReqType       reflect.Type
	RespType      reflect.Type
}
