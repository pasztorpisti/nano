# tests

## Intro

This is a set of quick and dirty integration tests for the servers.

These tests will compile the severs found in the `nano/examples/example1/servers`
directory to linux platform and then run them in docker containers to be able to
execute a very primitive integration test against them.

## Requirements

I've successfully tried the scripts on OS X and Debian Linux.

Running these tests requires only a docker installation on your machine. Make
sure that the user that runs the test scripts has docker access (tip: make sure
the user is added to the docker group).

No golang installation is needed on your machine because the `go build` is
executed inside a container with the correct go version and environment. For
this reason it's enough to clone this repo somewhere on your machine and run
one of the `test_server1.sh`, `test_server2.sh` and `test_server3.sh` scripts.

A lot of people use docker-compose or a similar tool to do some of the things
done by my scripts but I've deliberately chosen to depend only on docker for
this overly simple example.
