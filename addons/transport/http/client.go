package http

import (
	"bytes"
	"io"
	"net/http"
	"reflect"

	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/discovery"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
	"github.com/pasztorpisti/nano/addons/transport/http/serialization"
	"github.com/pasztorpisti/nano/addons/util"
)

type ClientOptions struct {
	Client        *http.Client
	Discoverer    discovery.Discoverer
	Serializer    *serialization.ClientSideSerializer
	PrefixURLPath bool
}

// DefaultClientOptions is used by NewClient when its opts parameter is nil.
// If DefaultClientOptions is nil then the opts parameter of NewClient can't be nil.
var DefaultClientOptions *ClientOptions

func NewClient(opts *ClientOptions, cfg *config.ServiceConfig) nano.Service {
	if opts == nil {
		if DefaultClientOptions == nil {
			panic("both opts and DefaultClientOptions are nil")
		}
		opts = DefaultClientOptions
	}

	endpoints := make(map[reflect.Type]*config.EndpointConfig, len(cfg.Endpoints))
	for _, ep := range cfg.Endpoints {
		if _, ok := endpoints[ep.ReqType]; ok {
			panic("multiple endpoints have the same req type: " + ep.ReqType.String())
		}
		endpoints[ep.ReqType] = ep
	}

	return &client{
		svcName:   cfg.ServiceName,
		endpoints: endpoints,
		opts:      opts,
	}
}

func (p *ClientOptions) client() *http.Client {
	if p.Client == nil {
		return http.DefaultClient
	}
	return p.Client
}

// client implements the nano.Service interface.
type client struct {
	svcName   string
	endpoints map[reflect.Type]*config.EndpointConfig
	opts      *ClientOptions
}

func (p *client) Name() string {
	return p.svcName
}

func (p *client) Init(cs nano.ClientSet) error {
	return nil
}

func (p *client) Handle(c *nano.Ctx, req interface{}) (resp interface{}, err error) {
	reqType := reflect.TypeOf(req)
	if reqType.Kind() != reflect.Ptr {
		return nil, p.Err(nil, "expected a pointer request type, got "+reqType.String())
	}
	ec, ok := p.endpoints[reqType.Elem()]
	if !ok {
		return nil, p.Err(nil, "can't handle request type "+reqType.String())
	}

	addr, err := p.opts.Discoverer.Discover(p.svcName)
	if err != nil {
		return nil, err
	}
	url := "http://" + addr
	if p.opts.PrefixURLPath {
		url += "/" + p.svcName
	}
	url += ec.Path

	var reqBody io.Reader
	var header http.Header
	h, body, err := p.opts.Serializer.SerializeRequest(ec, c, req)
	if err != nil {
		return nil, p.Err(err, "error serializing request")
	}
	reqBody = bytes.NewReader(body)
	header = h

	httpReq, err := http.NewRequest(ec.Method, url, reqBody)
	if err != nil {
		return nil, p.Err(err, "error creating request")
	}
	httpReq.Header = header

	httpResp, err := p.opts.client().Do(httpReq)
	if err != nil {
		return nil, p.Err(err, "http request failure")
	}
	defer httpResp.Body.Close()

	respObj, respErr, err := p.opts.Serializer.DeserializeResponse(ec, c, httpResp)
	if err != nil {
		return nil, p.Err(err, "error deserializing response")
	}
	return respObj, respErr
}

func (p *client) Err(cause error, msg string) error {
	return util.Err(cause, "service "+p.svcName+": "+msg)
}
