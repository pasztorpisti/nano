package nano

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

// BadReqTypeError should be returned by the Service.Handle method when it
// receives a req object of type it can't handle.
var BadReqTypeError = errors.New("bad request type")

// RunServer takes a set of initialised services and a list of listeners and
// initialises the listeners and then listens with them. Blocks and returns only
// when all listeners returned.
//
// It panics if the Init of a listener returns an error. Terminates the server
// process if the Listen method of a listener returns with an error.
var RunServer = func(ss ServiceSet, listeners ...Listener) {
	runServer(func(l Listener, err error) {
		log.Println("Listener failure :: " + err.Error())
		os.Exit(1)
	}, ss, listeners...)
}

var runServer = func(onError func(Listener, error), ss ServiceSet, listeners ...Listener) {
	for _, listener := range listeners {
		err := listener.Init(ss)
		if err != nil {
			panic("Error initialising listener :: " + err.Error())
		}
	}

	var wg sync.WaitGroup
	for _, listener := range listeners {
		wg.Add(1)
		listener := listener
		go func() {
			defer wg.Done()
			err := listener.Listen()
			if err != nil {
				onError(listener, err)
			}
		}()
	}
	wg.Wait()
}

// NewServiceSet creates a new ServiceSet object from the given services.
// The creation of the ServiceSet involves initialising the services before this
// function returns. The initialisation of the services includes the resolution
// of dependencies between each other by obtaining Client interfaces to each
// other in their Init methods.
var NewServiceSet = func(services ...Service) ServiceSet {
	ss := make(serviceSet, len(services))
	for _, svc := range services {
		ss[svc.Name()] = svc
	}

	for _, svc := range services {
		if serviceInit, ok := svc.(ServiceInit); ok {
			err := serviceInit.Init(NewClientSet(ss, svc.Name()))
			if err != nil {
				panic(fmt.Sprintf("error initialising service %q :: %v", svc.Name(), err))
			}
		}
	}

	for _, svc := range services {
		if serviceInitFinished, ok := svc.(ServiceInitFinished); ok {
			err := serviceInitFinished.InitFinished()
			if err != nil {
				panic(fmt.Sprintf("InitFinished error in service %q :: %v", svc.Name(), err))
			}
		}
	}

	return ss
}

// NewTestClientSet is a convenience helper for tests to wrap a set of services
// into a ServiceSet and then into a ClientSet in one step.
var NewTestClientSet = func(services ...Service) ClientSet {
	return NewClientSet(NewServiceSet(services...), "test")
}

// NewClientSet creates a ClientSet object from the given ServiceSet for the
// owner identified by ownerName. The ownerName can be anything but it should
// be the name of the service if the ClientSet is created for a service.
//
// The returned ClientSet belongs to the owner specified by ownerName and
// requests made through the Client objects returned by this ClientSet will send
// the ownerName to the called services which can inspect ownerName in their
// request context in Ctx.ClientName.
//
// You won't need this in your services/servers/tests. This function is public
// only to make nano hackable for experiments.
// You can replace this function with your own implementation to change the
// ClientSet implementation passed by NewServiceSet to ServiceInit.Init.
var NewClientSet = func(ss ServiceSet, ownerName string) ClientSet {
	return &clientSet{
		ss:        ss,
		ownerName: ownerName,
	}
}

// NewClient creates a new Client object that can be used to send requests to
// the svc service. The ownerName can be anything but it should be the name of
// the service if the Client is created for a service.
//
// The returned Client belongs to the owner specified by ownerName and requests
// made through the Client object will send the ownerName to the called services
// which can inspect ownerName in their request context in Ctx.ClientName.
//
// You won't need this in your services/servers/tests. This function is public
// only to make nano hackable for experiments.
// You can replace this function with your own implementation to change the
// Client implementation returned by ClientSet.
var NewClient = func(svc Service, ownerName string) Client {
	return &client{
		svc:       svc,
		ownerName: ownerName,
	}
}

// GeneratedReqIDBytesLen is the number of random bytes included in newly
// generated RequestIDs returned by the default implementation of NewReqID.
// Note that NewReqID returns a string that is the hex representation of the
// random generated ReqID byte array so the length of the returned ReqID string
// is the double of this value.
var GeneratedReqIDBytesLen = 16

// NewReqID generates a new random request ID. The default implementation
// generates a random byte array of length defined by GeneratedReqIDBytesLen and
// converts the byte array into a hex string.
//
// You won't need this in your services/servers/tests. This function is public
// only to make nano hackable for experiments.
var NewReqID = func() (string, error) {
	reqIDBytes := make([]byte, GeneratedReqIDBytesLen)
	_, err := rand.Read(reqIDBytes)
	if err != nil {
		return "", fmt.Errorf("error generating request ID :: %v", err)
	}
	return fmt.Sprintf("%x", reqIDBytes), nil
}

// NewContext has to return a new context.Context object and a related
// cancel func for a new request. The returned cancel func is guaranteed to be
// called exactly once, it doesn't have to handle multiple calls like the cancel
// funcs returned by the standard context package.
//
// The clientContext parameter might be nil if the initiator of the request
// isn't a service (e.g.: test) or if there is a transport layer between the
// caller and the called service. If both services reside in the same server
// executable then clientContext isn't nil but even in this case it is a
// bad idea to make clientContext the parent of the newly created context
// because that behavior would be very different from the scenario in which
// there is a transport layer between the services. However it completely makes
// sense to propagate the deadline and the cancellation of clientContext to the
// newly created context.
//
// You won't need this in your services/servers/tests. This function is public
// only to make nano hackable for experiments.
var NewContext = func(clientContext context.Context) (context.Context, context.CancelFunc) {
	c := context.Background()
	if clientContext == nil {
		return context.WithCancel(c)
	}

	var cancel context.CancelFunc
	if deadline, ok := clientContext.Deadline(); ok {
		c, cancel = context.WithDeadline(c, deadline)
	} else {
		c, cancel = context.WithCancel(c)
	}

	cancelChan := make(chan struct{})
	go func() {
		select {
		case <-clientContext.Done():
		case <-cancelChan:
		}
		cancel()
	}()
	return c, func() { close(cancelChan) }
}

// serviceSet implements the ServiceSet interface.
type serviceSet map[string]Service

func (p serviceSet) LookupService(svcName string) (Service, error) {
	svc, ok := p[svcName]
	if !ok {
		return nil, fmt.Errorf("service not found: %v", svcName)
	}
	return svc, nil
}

// clientSet implements the ClientSet interface.
type clientSet struct {
	ss        ServiceSet
	ownerName string
}

func (p *clientSet) LookupClient(svcName string) Client {
	svc, err := p.ss.LookupService(svcName)
	if err != nil {
		panic(fmt.Sprintf("service %q failed to lookup client %q :: %v",
			p.ownerName, svcName, err))
	}
	return NewClient(svc, p.ownerName)
}

// client implements the Client interface.
type client struct {
	svc       Service
	ownerName string
}

func (p *client) Request(c *Ctx, req interface{}) (resp interface{}, err error) {
	var c2 Ctx
	if c != nil {
		c2 = *c
	}

	if c2.ReqID == "" {
		if c2.ReqID, err = NewReqID(); err != nil {
			return nil, err
		}
	}

	var cancel context.CancelFunc
	c2.Context, cancel = NewContext(c2.Context)
	defer cancel()

	c2.Svc, c2.ClientName = p.svc, p.ownerName
	return p.svc.Handle(&c2, req)
}
