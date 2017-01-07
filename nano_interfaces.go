package nano

import "context"

// Listener is a component in the server executable that can receive requests
// from the outside world and forward them to the already initialised services
// located inside the server executable.
type Listener interface {
	// Init is called with the initialised services of the server.
	Init(ServiceSet) error

	// Listen is called after Init. This method is called on its own goroutine,
	// this is how a server can listen on more than one listeners in parallel.
	Listen() error
}

// ServiceSet is a set of initialised services. During initialisation the
// services have already resolved the dependencies between each other.
type ServiceSet interface {
	// LookupService returns the service with the specified name.
	LookupService(svcName string) (Service, error)
}

// Service is an interface that has to be implemented by all services.
type Service interface {
	// Name returns the name of the service. I recommend using simple names with
	// safe characters, snake_case is a very good choice. If you use simple
	// names then it will be easy to use the same name for directories, packages
	// host names, identifiers in code, etc...
	Name() string

	// Handle is the handler function of the service. It can receive any type
	// of request object and respond with any type of response.
	// If the type of the req parameter can't be handled by Handler then it
	// it should return a nano.BadReqTypeError.
	// Note that nil is a valid response value if you define the API of your
	// service that way.
	//
	// While there are no restrictions on the type of request and response
	// objects it is wise to use easy-to-serialize types (e.g.: json/protobuf
	// serializable types) in order to make it easy to transfer req/resp objects
	// through the wire. I recommend using struct pointers.
	Handle(c *Ctx, req interface{}) (resp interface{}, err error)
}

// ServiceInit is an interface that can optionally be implemented by a Service
// object. If a service implements this interface then it receives an Init call
// when the ServiceSet that contains this service is being created.
type ServiceInit interface {
	// Init is called when the ServiceSet is being created. This is the right
	// time to resolve the dependencies of this service by asking for the
	// clients of other services from the received ClientSet. It is discouraged
	// to store the received ClientSet for later use - use it only before
	// returning from Init.
	//
	// Note that when your Init is called other services might be uninitialised
	// so you shouldn't send requests to them through the clients your obtained
	// using the ClientSet parameter. If you have to send a request to
	// another service at init time then you should implement the
	// ServiceInitFinished interface.
	//
	// Init is guaranteed to be called before the InitFinished and Handler
	// methods of any service in the ServiceSet that is being created.
	Init(ClientSet) error
}

// ServiceInitFinished is an interface that can optionally be implemented by a
// Service object. If a service implements this interface then it receives an
// InitFinished call when the ServiceSet that contains this service is being
// created.
type ServiceInitFinished interface {
	// InitFinished is called during the creation of a ServiceSet only after
	// calling the Init of all services in the ServiceSet.
	//
	// InitFinished is guaranteed to be called after the Init of your service
	// but it isn't guaranteed to be called before the Handler method because
	// the InitFinished of other services might be called earlier than your
	// InitFinished and they might send requests to your service.
	InitFinished() error
}

// ClientSet can be used by a service to obtain client interfaces to other
// services.
type ClientSet interface {
	// LookupClient returns a client interface to the service identified by the
	// given svcName. Trying to lookup a service that doesn't exist results in
	// a panic. This is by design to make the initialisation code simpler.
	// Normally services lookup their client at server startup in their Init
	// methods where a panic is acceptable.
	LookupClient(svcName string) Client
}

// Client is an interface that can be used to send requests to a service.
type Client interface {
	// Requests sends a request to the service targeted by this client object.
	// The c parameter is the context of the caller service but it is allowed
	// to be nil or a Ctx instance with only some of its fields set.
	//
	// Note the similarity between the signature of Client.Request and
	// Service.Handle. The reason for calling another service through a Client
	// interface instead of directly calling its Service.Handle method is that
	// the request context (the c parameter) isn't directly passed between
	// services. While all other parameters and return values are transferred
	// directly between Client.Request and Service.Handle a new request context
	// (*Ctx) has to be created when the control is transferred to another
	// service. While it might make sense to directly copy and pass some fields
	// of the request context of the caller (e.g.: Ctx.ReqID), most other fields
	// of the request context have to be adjusted (e.g.: Ctx.Svc) that is done
	// by this Client interface.
	//
	// To clarify things: The c parameter of Client.Request is the request
	// context of the caller, while the c parameter of the Service.Handler will
	// be another request context newly created for the called service.
	Request(c *Ctx, req interface{}) (resp interface{}, err error)
}

// Ctx holds context data for a given request being served by a service.
// If you fork and modify this repo (or roll your own stuff) for a project then
// one of the benefits is being able to specify the contents of your request
// context structure that is being passed around in your service functions.
// I've included a few sensible fields but your project might want to exclude
// some of these or perhaps add new ones. An example: If you implement
// authentication and authorization at the transport layer then you can use a
// field in the Ctx structure to pass user id info to the business logic if it
// makes sense in the given project.
type Ctx struct {
	// ReqID is the ID of the request that entered the cluster. Note that by
	// making calls to other services by calling Client.Request and passing this
	// context info the request inside the other service will copy this ReqID.
	// For this reason ReqID works as a correlation ID between the requests
	// handled by a cluster of your services and can be used for example to
	// track logs.
	ReqID string

	// Context can be useful when you call some std library functions that
	// accept a context.
	Context context.Context

	// Svc is the current service.
	Svc Service

	// ClientName is the name of the entity that initiated the request. It is
	// usually the name of another service but it can be anything else, for
	// example "test" if the request has been initiated by a test case.
	ClientName string
}

// WithContext returns a shallow copy of the context after assigning the given
// ctx to the Context field.
func (c *Ctx) WithContext(ctx context.Context) *Ctx {
	if ctx == nil {
		panic("nil context")
	}
	c2 := *c
	c2.Context = ctx
	return &c2
}
