package static

import "github.com/pasztorpisti/nano/addons/discovery"

// Discoverer implements the discovery.Discoverer interface. It can resolve
// a predefined set of service names into predefined net.Dial compatible
// addresses.
type Discoverer map[string]string

func (v Discoverer) Discover(name string) (string, error) {
	if addr, ok := v[name]; ok {
		return addr, nil
	}
	return "", discovery.NotFoundError
}
