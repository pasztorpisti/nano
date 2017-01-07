package test

import (
	"flag"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/util"
)

func Init() {
	flag.Parse()

	origNewClient := nano.NewClient
	nano.NewClient = func(svc nano.Service, ownerName string) nano.Client {
		return errorFilterClient{origNewClient(svc, ownerName)}
	}
}

// errorFilterClient implements nano.Client. When we execute business logic
// tests all services (or a combination of services and mock services) reside
// in the same test executable. For this reason services can return error
// objects of any type without problems. However when we place the services into
// separate server executables and put a transport layer between them the
// transport layer will support only a few error types and can't return other
// object types. This can cause different behavior between tests and server
// executables that can lead to successful tests and code that can't detect
// errors correctly when services are called through transport layer between
// two servers.
//
// To avoid the previous scenario we hook the nano.NewClient in this test config
// package and wrap the client returned by the original nano.NewClient function
// into an instance of this client. This client makes sure that during tests
// errors are returned the same way as our transport implementation returns them
// through network. We use the github.com/pasztorpisti/nano/addons/transport/http
// transport that supports only the NanoError error type and always returns nil
// or NanoError. This client simulates this behavior.
type errorFilterClient struct {
	client nano.Client
}

func (p errorFilterClient) Request(c *nano.Ctx, req interface{}) (interface{}, error) {
	resp, err := p.client.Request(c, req)
	if err != nil {
		return nil, util.ErrCode(nil, util.GetErrCode(err), err.Error())
	}
	return resp, nil
}
