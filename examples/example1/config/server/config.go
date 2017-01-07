package server

import (
	"fmt"

	"github.com/pasztorpisti/nano/addons/transport/http"
	"github.com/pasztorpisti/nano/addons/transport/http/serialization/json"
	"github.com/pasztorpisti/nano/examples/example1/config/common"
)

func Init() {
	common.Init()

	http.DefaultListenerOptions = &http.ListenerOptions{
		BindAddr:      fmt.Sprintf(":%d", common.Port),
		Serializer:    json.ServerSideSerializer,
		PrefixURLPath: true,
	}
}
