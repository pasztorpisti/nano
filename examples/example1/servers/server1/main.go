package main

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc2"
	"github.com/pasztorpisti/nano/examples/example1/config/server"
	svc_svc1 "github.com/pasztorpisti/nano/examples/example1/services/svc1"
	svc_svc2 "github.com/pasztorpisti/nano/examples/example1/services/svc2"
	svc_svc3 "github.com/pasztorpisti/nano/examples/example1/services/svc3"
	svc_svc4 "github.com/pasztorpisti/nano/examples/example1/services/svc4"
)

func main() {
	server.Init()

	ss := nano.NewServiceSet(
		svc_svc1.New(),
		svc_svc2.New(),
		svc_svc3.New(),
		svc_svc4.New(),
	)
	listener := http.NewListener(nil, svc2.HTTPTransportConfig)
	nano.RunServer(ss, listener)
}
