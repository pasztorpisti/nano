package http

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"github.com/pasztorpisti/nano"
	"github.com/pasztorpisti/nano/addons/log"
	"github.com/pasztorpisti/nano/addons/transport/http/config"
	"github.com/pasztorpisti/nano/addons/transport/http/serialization"
	"github.com/pasztorpisti/nano/addons/util"
)

type ListenerOptions struct {
	BindAddr      string
	Serializer    *serialization.ServerSideSerializer
	PrefixURLPath bool
}

var DefaultListenerOptions *ListenerOptions

func NewListener(opts *ListenerOptions, cfgs ...*config.ServiceConfig) nano.Listener {
	if opts == nil {
		if DefaultListenerOptions == nil {
			panic("both opts and DefaultListenerOptions are nil")
		}
		opts = DefaultListenerOptions
	}
	return &listener{
		cfgs: cfgs,
		opts: opts,
	}
}

type listener struct {
	cfgs   []*config.ServiceConfig
	opts   *ListenerOptions
	router *httprouter.Router
}

func (p *listener) Init(srv nano.ServiceSet) error {
	p.router = httprouter.New()
	p.router.GET("/health-check", func(http.ResponseWriter, *http.Request, httprouter.Params) {})

	duplicateCheck := map[string]struct{}{}
	for _, cfg := range p.cfgs {
		svc, err := srv.LookupService(cfg.ServiceName)
		if err != nil {
			return util.Err(err, "listener couldn't lookup a service")
		}

		for _, ec := range cfg.Endpoints {
			path := ec.Path
			if p.opts.PrefixURLPath {
				path = "/" + cfg.ServiceName + path
			}
			id := ec.Method + " " + path
			if _, ok := duplicateCheck[id]; ok {
				return fmt.Errorf("service %v: duplicate endpoint: %v",
					cfg.ServiceName, id)
			}
			duplicateCheck[id] = struct{}{}

			ep := &endpoint{
				cfg:        ec,
				svc:        svc,
				Serializer: p.opts.Serializer,
			}
			p.router.Handle(ec.Method, path, ep.Handler)
		}
	}

	return nil
}

func (p *listener) Listen() error {
	return http.ListenAndServe(p.opts.BindAddr, p.router)
}

type endpoint struct {
	cfg        *config.EndpointConfig
	svc        nano.Service
	Serializer *serialization.ServerSideSerializer
}

func (p *endpoint) Handler(w http.ResponseWriter, r *http.Request,
	rp httprouter.Params) {
	req, ri, err := p.Serializer.DeserializeRequest(p.cfg, r)
	if err != nil {
		log.Err(nil, err, "error deserialising request")
		err = p.Serializer.SerializeResponse(p.cfg, nil, w, r, nil, err)
		if err != nil {
			log.Err(nil, err, "error serialising error response")
		}
		return
	}

	client := nano.NewClient(p.svc, ri.ClientName)

	c := &nano.Ctx{
		ReqID: ri.ReqID,
	}
	resp, err := client.Request(c, req)

	if err != nil {
		resp = nil
	} else {
		expectedType := p.cfg.RespType
		if expectedType != nil {
			expectedType = reflect.PtrTo(expectedType)
		}
		if reflect.TypeOf(resp) != expectedType {
			// this is a programming error in the service
			log.Errf(c, err, "service returned an object of type %v, want %v",
				reflect.TypeOf(resp), expectedType)
			err = p.Serializer.SerializeResponse(p.cfg, c, w, r, nil, serverError)
			if err != nil {
				log.Err(c, err, "error serialising error response")
			}
			return
		}
	}

	err = p.Serializer.SerializeResponse(p.cfg, c, w, r, resp, err)
	if err != nil {
		log.Err(c, err, "error serialising response")
		return
	}
}

var serverError = util.ErrCode(nil, config.ErrorCodeServerError,
	"Internal Server Error")
