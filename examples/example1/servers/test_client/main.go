package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/discovery/static"
	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/examples/example1/api_go/svc2"
	"github.com/pasztorpisti/nano/examples/example1/config/client"
	"github.com/pasztorpisti/nano/examples/example1/config/common"
)

var addr = flag.String("addr", fmt.Sprintf("localhost:%d", common.Port),
	"addr:port of the server to connect to")
var reqType = flag.String("req", "Req",
	"The type of request to perform. Must be 'Req' or 'GetReq'.")

func main() {
	client.Init()
	http.DefaultClientOptions.Discoverer = static.Discoverer{"svc2": *addr}

	ss := nano.NewTestClientSet(
		http.NewClient(nil, svc2.HTTPTransportConfig),
	)
	svc2Client := ss.LookupClient("svc2")

	var req interface{}
	switch strings.ToLower(*reqType) {
	case "req":
		req = &svc2.Req{Param: "param"}
	case "getreq":
		req = &svc2.GetReq{}
	default:
		fmt.Printf("Invalid request type: %q\n", *reqType)
		os.Exit(1)
	}
	fmt.Printf("req=%#v\n", req)

	resp, err := svc2Client.Request(nil, req)

	fmt.Printf("resp=%#v err=%v\n", resp, err)

	if err != nil {
		os.Exit(1)
	}
}
