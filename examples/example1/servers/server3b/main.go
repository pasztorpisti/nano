package main

import (
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc1"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc2"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc3"
	"github.com/pasztorpisti/nano/examples/example1/config/client"
	"github.com/pasztorpisti/nano/examples/example1/config/server"
	svc_svc2 "github.com/pasztorpisti/nano/examples/example1/services/svc2"
)

func main() {
	client.Init()
	server.Init()

	ss := nano.NewServiceSet(
		http.NewClient(nil, svc1.HTTPTransportConfig),
		svc_svc2.New(),
		http.NewClient(nil, svc3.HTTPTransportConfig),
	)
	listener := http.NewListener(nil, svc2.HTTPTransportConfig)
	nano.RunServer(ss, listener)
}
