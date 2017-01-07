package main

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc3"
	"github.com/pasztorpisti/nano/examples/example1/config/server"
	svc_svc3 "github.com/pasztorpisti/nano/examples/example1/services/svc3"
	svc_svc4 "github.com/pasztorpisti/nano/examples/example1/services/svc4"
)

func main() {
	server.Init()

	ss := nano.NewServiceSet(
		svc_svc3.New(),
		svc_svc4.New(),
	)
	listener := http.NewListener(nil, svc3.HTTPTransportConfig)
	nano.RunServer(ss, listener)
}
