/*
Package discovery provides different ways to resolve the name of your services
to their respective "addr:port".

Different implementations of the Discoverer interface have been placed to
separate sub-packages intentionally. This way the unused implementations don't
become dependencies of the compiled server executable.
*/
package discovery
