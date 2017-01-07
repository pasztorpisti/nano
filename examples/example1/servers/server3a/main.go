package main

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc1"
	"github.com/pasztorpisti/nano/examples/example1/config/server"
	svc_svc1 "github.com/pasztorpisti/nano/examples/example1/services/svc1"
)

func main() {
	server.Init()

	ss := nano.NewServiceSet(
		svc_svc1.New(),
	)
	listener := http.NewListener(nil, svc1.HTTPTransportConfig)
	nano.RunServer(ss, listener)
}
