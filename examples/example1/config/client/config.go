package client

import (
	"fmt"
	net_http "net/http"

	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/addons/transport/http/serialization/json"
	"github.com/pasztorpisti/nano/examples/example1/config/common"
)

func Init() {
	common.Init()

	http.DefaultClientOptions = &http.ClientOptions{
		Client:        net_http.DefaultClient,
		Discoverer:    discoverer{},
		Serializer:    json.ClientSideSerializer,
		PrefixURLPath: true,
	}
}

type discoverer struct{}

func (discoverer) Discover(name string) (string, error) {
	return fmt.Sprintf("%s:%d", name, common.Port), nil
}
