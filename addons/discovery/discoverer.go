package discovery

import "errors"

// NotFoundError can be returned by returned by Discoverer.Discover when it
// can't locate a given service.
var NotFoundError = errors.New("not found")

// Discoverer helps locating services.
type Discoverer interface {
	// Discover receives the name of a service and returns the "host:port" of
	// the service in a format expected by net.Dial.
	//
	// It should return NotFoundError if the given service was not found but
	// returning other errors is also allowed. E.g.: If a network error happens
	// during a DNS lookup then returning the network error instead of a
	// NotFoundError makes sense.
	Discover(name string) (string, error)
}
