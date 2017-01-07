package common

import "flag"

const (
	// Port is the predefined fix port on which our servers will listen.
	// Since we run each service in its own container they have their own IP
	// address and a hostname that is the same as the service name.
	Port = 8000
)

var alreadyInitialised = false

// Init contains initialisation code that is shared by client and server config.
func Init() {
	if alreadyInitialised {
		return
	}
	alreadyInitialised = true

	flag.Parse()
}
