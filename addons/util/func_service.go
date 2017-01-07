package util

import "github.com/pasztorpisti/nano"

// HandlerFunc is a function type compatible with the Service.Handle interface
// method.
type HandlerFunc func(c *nano.Ctx, req interface{}) (resp interface{}, err error)

// ServiceOpts is used as an incoming parameter for the NewServiceOpts function.
type ServiceOpts struct {
	// Name is the name of the service object to create.
	Name string

	// Handler is the handler function of the service object to create.
	Handler HandlerFunc

	// Init contains code for the Init method of the service object. Can be nil.
	Init func(cs nano.ClientSet) error

	// InitFinished contains code for the InitFinished method of the service
	// object. Can be nil.
	InitFinished func() error
}

// NewService creates a service object from the given service name and handler
// function. This function is useful for creating simple mock service objects
// inside test functions but its usage isn't restricted to tests.
func NewService(name string, handler HandlerFunc) nano.Service {
	return NewServiceOpts(ServiceOpts{
		Name:    name,
		Handler: handler,
	})
}

// NewServiceOpts creates a service object from the given service name and
// handler function(s). This function is useful for creating simple mock service
// objects inside test functions but its usage isn't restricted to tests.
func NewServiceOpts(opts ServiceOpts) nano.Service {
	if opts.Name == "" {
		panic("name is an empty string")
	}
	if opts.Handler == nil {
		panic("handle is nil")
	}

	svc := service{
		name:    opts.Name,
		handler: opts.Handler,
	}

	switch {
	case opts.Init != nil && opts.InitFinished != nil:
		return &struct {
			service
			serviceInit
			serviceInitFinished
		}{
			service:             svc,
			serviceInit:         opts.Init,
			serviceInitFinished: opts.InitFinished,
		}
	case opts.Init != nil:
		return &struct {
			service
			serviceInit
		}{
			service:     svc,
			serviceInit: opts.Init,
		}
	case opts.InitFinished != nil:
		return &struct {
			service
			serviceInitFinished
		}{
			service:             svc,
			serviceInitFinished: opts.InitFinished,
		}
	default:
		return &svc
	}
}

// service implements the nano.Service interface.
type service struct {
	name    string
	handler HandlerFunc
}

func (p *service) Name() string {
	return p.name
}

func (p *service) Handle(c *nano.Ctx, req interface{}) (resp interface{}, err error) {
	return p.handler(c, req)
}

// serviceInit implements the nano.ServiceInit interface.
type serviceInit func(cs nano.ClientSet) error

func (f serviceInit) Init(cs nano.ClientSet) error {
	return f(cs)
}

// serviceInitFinished implements the nano.ServiceInitFinished interface.
type serviceInitFinished func() error

func (f serviceInitFinished) InitFinished() error {
	return f()
}
