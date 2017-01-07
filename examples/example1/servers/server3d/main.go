package main

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc4"
	"github.com/pasztorpisti/nano/examples/example1/config/server"
	svc_svc4 "github.com/pasztorpisti/nano/examples/example1/services/svc4"
)

func main() {
	server.Init()

	ss := nano.NewServiceSet(
		svc_svc4.New(),
	)
	listener := http.NewListener(nil, svc4.HTTPTransportConfig)
	nano.RunServer(ss, listener)
}
