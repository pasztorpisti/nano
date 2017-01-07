# nano

Transport-agnostic testing-friendly nano-framework for go micro-services.

# Short Intro

- Separation between business logic and lower level layers (transport, discovery).
- A service is a unit of business logic that can manage a few types of messages.
- Units of business logic (services) are easy to compile into a single or
  multiple server executables.
- The development and unit/integration testing of units of business logic
  (services) can be done with the standard go test framework:
  - No need to compile and start server executables
  - No need for an infrastructure
  - Mocking a service can be done by writing a mock handler function

# Long Intro: Goals and Principles

This lightweight microservice framework has the following ambitious goals:

- A service should be a unit of business logic that doesn't contain any platform
  or infrastructure related details like serialization, transport, discovery.
  A service has a name and its most important part which is the request
  handler function (that isn't a HTTP handler func).
- Services (service handler functions) should be able to access each other
  easily either through transport (like HTTP or message queues) or simplified
  in-memory communication when services reside in the same (test or server)
  executable. A service shouldn't know whether it communicates through transport
  or in-memory channel.
- A standard golang test should be able to include and connect the business
  logic of several golang services into a single test executable in order to
  avoid writing server executables and setting up an infrastructure for testing.
  Result: extremely comfortable and rapid development of business logic with
  the standard go test framework.
- A golang test should be able to mock one or more services in a service
  integration test in the simplest possible way. E.g.: by providing mock service
  handler functions. This is orders of magnitude easier than implementing mock
  server executables and starting them up in a test infrastructure.
  Test execution is also much faster.
- After writing the services (business logic) and their tests a developer should
  be able to include the services into server executables and connect them with
  a transport layer (HTTP, msg queue) very easily. A server executable (like a
  test) should be able to include all services or an arbitrary subset of them.
  This feels like a greyscale between monolith and microservice architecture.
- Classic server/infrastructure integration (or e2e) tests are still necessary
  but you need less. At the same time being able to develop business logic
  without infrastructure tests and being able to include all services into
  a single simple executable can be a huge speed boost for development.

The **nano** package is a minimalistic implementation of the previously described
ideas. It is ~200 lines of code without comments and ~400 with comments
(see [nano_interfaces.go](nano_interfaces.go) and [nano.go](nano.go)). These
figures don't include tests and addons.
Addons (including transport) integrate transparently without the nano package
knowing about them. You'll see how through a few simple examples.

Feel free to use nano in your experiments. In case of a serious project I
recommend forking nano and tailoring it to your needs or using it as a source
of inspiration and ideas while implementing your own.

I've used go 1.7 for development and haven't tested other go versions.

# Workflow

This section introduces the nano framework by guiding you through the
development of a very simple example microservice cluster.

## 1a. Implementing services

In nano a service is an implementation of the `nano.Service` interface.
A service can additionally implement the optional `nano.SerivceInit` and/or
`nano.ServiceInitFinished` interfaces:

```go
type Service interface {
	Name() string
	Handle(c *Ctx, req interface{}) (resp interface{}, err error)
}

type ServiceInit interface {
	Init(ClientSet) error
}

type ServiceInitFinished interface {
	InitFinished() error
}

```
See the the above interfaces with their comments in
[nano_interfaces.go](nano_interfaces.go).

A simple example service can be found
[here](examples/example1/services/svc4/svc.go) and a bit more complex service
that implements `nano.ServiceInit` to resolve its dependencies
[here](examples/example1/services/svc3/svc.go).

I recommend using short [snake_case](https://en.wikipedia.org/wiki/Snake_case)
service names because in that case you can consistently use the same name for
directory names, host names, package names, source code identifiers, etc...

Each service has one or more request-response message pairs that are handled
by the handler func of the service. It is recommended to use easy to serialise
message types to make the addition of the transport layer easy in a later stage.
I recommend using struct pointers. In my examples I've generated my request and
response structs from protobuf definitions but you can define these structs
manually if you prefer that (for example for json serialization).

Microservices sometimes aren't only about golang so it is a good idea to
maintain a language independent description of the API. This is why the
[`api` directory](examples/example1/api) exists in my project structure with
protobuf definitions.
I'm using the [`api_go/generate.sh` script](examples/example1/api_go/generate.sh)
to generate the go-specific stuff from it. The implementation of the service
comes after defining the request-response type pairs of the service in the API.
Note that the `api` directory contains some other things too
(e.g.: `http_transport_config.json`) that aren't needed in the first stage while
implementing the service. Please ignore them.

## 1b. Testing services

After writing one or more services you want to test each service as a unit and
perhaps integrating several services together surrounded by mock services if
necessary. Since the service interface is simple and well defined by the nano
package you can actually write tests before the service code but I'm a
test-after guy. :-)

This framework allows the creation and integration of multiple services by
creating a `nano.ServiceSet`. Creating a `nano.ServiceSet` initialises the
included services and resolves the dependencies between them. After creation
you can send requests to any of the services in the set. Note that the
`nano.ServiceSet` is a universal object that is used not only while writing
tests during development but also for containing the services inside server
executables.

This is a newly created and initialised `nano.ServiceSet` object:

```
+---------------------------------------+
|  ServiceSet                           |
|                                       |
|       +-----+  +-----+  +-----+       |
|       |     |  |     |  |     |       |
|       v     |  |     v  |     v       |
|  +------+ +-+--+-+ +----+-+ +------+  |
|  | svc1 | | svc2 | | svc3 | | svc4 |  |
|  +------+ +------+ +------+ +------+  |
|                                       |
+---------------------------------------+
"A -> B" means "A depends on B" or "A can send requests to B".
```

After the creation and initialisation of the `nano.ServiceSet` the tests can
interact with any of the services. Here is a test executable containing all
services integrated together without any network/transport between them:

```
+-------------------------------------------+
| test1                                     |
|                                           |
| +---------------------------------------+ |
| | ServiceSet                            | |
| |                                       | |
| |       +-----+  +-----+  +-----+       | |
| |       |     |  |     |  |     |       | |
| |       v     |  |     v  |     v       | |
| |  +------+ +-+--+-+ +----+-+ +------+  | |
| |  | svc1 | | svc2 | | svc3 | | svc4 |  | |
| |  +------+ +------+ +------+ +------+  | |
| |              ^                        | |
| |              |                        | |
| +--------------+------------------------+ |
|                |                          |
|          +-----+------+                   |
|          | svc2-test1 |                   |
|          +------------+                   |
+-------------------------------------------+
```

Source code:
[test1](examples/example1/services/svc2/tests/test1/svc_test.go)
[svc1](examples/example1/services/svc1/svc.go)
[svc2](examples/example1/services/svc2/svc.go)
[svc3](examples/example1/services/svc3/svc.go)
[svc4](examples/example1/services/svc4/svc.go)

Service interface definitions in protobuf (request/response pairs):
[svc1](examples/example1/api/svc1/requests.proto)
[svc2](examples/example1/api/svc2/requests.proto)
[svc3](examples/example1/api/svc3/requests.proto)
[svc4](examples/example1/api/svc4/requests.proto)

Note that a service can be replaced with a mock service easily while creating
the ServiceSet. Below you can see another test that uses mocks instead of
`svc1` and `svc3`. The `svc4` service has been left out because the mock version
of `svc3` doesn't depend on it:

```
+----------------------------------+
| test2                            |
|                                  |
| +------------------------------+ |
| | ServiceSet                   | |
| |                              | |
| |       +-----+  +-----+       | |
| |       |     |  |     |       | |
| |       v     |  |     v       | |
| |  +------+ +-+--+-+ +------+  | |
| |  | svc1 | | svc2 | | svc3 |  | |
| |  | mock | +------+ | mock |  | |
| |  +------+    ^     +------+  | |
| |              |               | |
| +--------------+---------------+ |
|                |                 |
|          +-----+------+          |
|          | svc2-test2 |          |
|          +------------+          |
+----------------------------------+
```

Source code:
[test2](examples/example1/services/svc2/tests/test2/svc_test.go)
[svc2](examples/example1/services/svc2/svc.go)

As you've seen it is very easy to write both unit tests and integration tests
to assert your business logic. All you need is nano and the go testing framework.
At this stage you don't have to implement server executables and you can avoid
setting up an infrastructure.

## 2. Including services into server executables 

The below example includes all services into a single server executable and
makes the `svc2` service accessible to the outside world:

```
+-------------------------------------------+
| server1                                   |
|                                           |
| +---------------------------------------+ |
| | ServiceSet                            | |
| |                                       | |
| |       +-----+  +-----+  +-----+       | |
| |       |     |  |     |  |     |       | |
| |       v     |  |     v  |     v       | |
| |  +------+ +-+--+-+ +----+-+ +------+  | |
| |  | svc1 | | svc2 | | svc3 | | svc4 |  | |
| |  +------+ +------+ +------+ +------+  | |
| |              ^                        | |
| |              |                        | |
| +--------------+------------------------+ |
|                |                          |
| +--------------+------------------------+ |
| | Listener     |                        | |
| |           +--+---+                    | |
| |           | svc2 |                    | |
| |           | cfg  |                    | |
| |           +------+                    | |
| +---------------------------------------+ |
+-------------------------------------------+
```

Source code:
[server1](examples/example1/servers/server1/main.go)

The above `server1` example includes all services into one server executable but
you are allowed to group your services into several server executables as you
wish. Below you can find the implementation of `server2a` and `server2b` that
contain the same services as the previously implemented `server1` but in the
"server2" example the `svc1` and `svc2` services reside in `server2a` while the
`svc3` and `svc4` services are in `server2b`.

```
+------------------------------------+     +-------------------------+
| server2a                           |     | server2b                |
|                                    |     |                         |
| +--------------------------------+ |     | +---------------------+ |
| | ServiceSet                     | |     | | ServiceSet          | |
| |                                | |     | |                     | |
| |       +-----+  +-----+         | |     | |     +-------+       | |
| |       |     |  |     |         | |     | |     |       |       | |
| |       v     |  |     v         | |     | |     |       v       | |
| |  +------+ +-+--+-+ +--------+  | |     | |  +--+---+ +------+  | |
| |  | svc1 | | svc2 | | svc3   |  | |     | |  | svc3 | | svc4 |  | |
| |  +------+ +------+ | client +--+-+--+  | |  +------+ +------+  | |
| |              ^     +--------+  | |  |  | |     ^               | |
| |              |                 | |  |  | |     |               | |
| +--------------+-----------------+ |  |  | +-----+---------------+ |
|                |                   |  |  |       |                 |
| +--------------+-----------------+ |  |  | +-----+---------------+ |
| | Listener     |                 | |  |  | |     |     Listener  | |
| |           +--+---+             | |  |  | |  +--+---+           | |
| |           | svc2 |             | |  |  | |  | svc3 |           | |
| |           | cfg  |             | |  +--+-+->| cfg  |           | |
| |           +------+             | |     | |  +------+           | |
| +--------------------------------+ |     | +---------------------+ |
+------------------------------------+     +-------------------------+
```

Sources code:
[server2a](examples/example1/servers/server2a/main.go)
[server2b](examples/example1/servers/server2b/main.go)
 
The `svc2` service accesses the `svc3` service through the `svc3 client`
service. The `svc3 client` service implementation is provided by the example
(HTTP) transport layer in the form of a generic client that we configure to
behave as a client for `svc3`.
 
The `svc3 client` is actually a "fake service" that implements the `nano.Service`
interface just like any other service but instead of containing business logic
it simply transfers the request and the response through network to communicate
with another server that contains the actual `svc3` implementation.
For this reason the nano framework doesn't even have to know that `svc3 client`
is actually a client.

The nano framework has only two requirements when it comes to implementing a
transport layer: The client-side of the transport has to implement the
`nano.Service` interface and has to provide the same interface as the service to
which it forwards the incoming requests through network. The server-side of the
transport has to implement the `nano.Listener` interface.

Note that the example HTTP transport implementation is an addon and you are
allowed and encouraged to implement your own transport layer(s) that suit the
needs of your project. E.g.: http transport with TLS, authentication or grpc
transport or msg queue transport, etc...

By taking it to the next level we can deploy each service in its own server executable:

```
+----------------+     +--------------------------------------+     +---------------------------+     +----------------+
| server3a       |     | server3b                             |     | server3c                  |     | server3d       |
|                |     |                                      |     |                           |     |                |
| +------------+ |     | +----------------------------------+ |     | +-----------------------+ |     | +------------+ |
| | ServiceSet | |     | | ServiceSet                       | |     | | ServiceSet            | |     | | ServiceSet | |
| |            | |     | |                                  | |     | |                       | |     | |            | |
| |            | |     | |       +-------+  +-------+       | |     | |     +---------+       | |     | |            | |
| |            | |     | |       |       |  |       |       | |     | |     |         |       | |     | |            | |
| |            | |     | |       v       |  |       v       | |     | |     |         v       | |     | |            | |
| |  +------+  | |     | |  +--------+ +-+--+-+ +--------+  | |     | |  +--+---+ +--------+  | |     | |  +------+  | |
| |  | svc1 |  | |     | |  | svc1   | | svc2 | | svc3   |  | |     | |  | svc3 | | svc4   |  | |     | |  | svc4 |  | |
| |  +------+  | |  +--+-+--+ client | +------+ | client +--+-+--+  | |  +------+ | client +--+-+--+  | |  +------+  | |
| |     ^      | |  |  | |  +--------+    ^     +--------+  | |  |  | |     ^     +--------+  | |  |  | |     ^      | |
| |     |      | |  |  | |                |                 | |  |  | |     |                 | |  |  | |     |      | |
| +-----+------+ |  |  | +----------------+-----------------+ |  |  | +-----+-----------------+ |  |  | +-----+------+ |
|       |        |  |  |                  |                   |  |  |       |                   |  |  |       |        |
| +-----+------+ |  |  | +----------------+-----------------+ |  |  | +-----+-----------------+ |  |  | +-----+------+ |
| |     |      | |  |  | |                |                 | |  |  | |     |                 | |  |  | |     |      | |
| |  +--+---+  | |  |  | |             +--+---+             | |  |  | |  +--+---+             | |  |  | |  +--+---+  | |
| |  | svc1 |  | |  |  | |             | svc2 |             | |  |  | |  | svc3 |             | |  |  | |  | svc4 |  | |
| |  | cfg  |<-+-+--+  | |             | cfg  |             | |  +--+-+->| cfg  |             | |  +--+-+->| cfg  |  | |
| |  +------+  | |     | |             +------+             | |     | |  +------+             | |     | |  +------+  | |
| | Listener   | |     | | Listener                         | |     | | Listener              | |     | | Listener   | |
| +------------+ |     | +----------------------------------+ |     | +-----------------------+ |     | +------------+ |
+----------------+     +--------------------------------------+     +---------------------------+     +----------------+
```
Source code:
[server3a](examples/example1/servers/server3a/main.go)
[server3b](examples/example1/servers/server3b/main.go)
[server3c](examples/example1/servers/server3c/main.go)
[server3d](examples/example1/servers/server3d/main.go)

## 3a. Deploying the servers

I highly recommend using containerisation technology for deployment both
on developer machines and in production (and any other stage between them in
your pipeline).

On a developer machine containers can be used as "lightweight virtual machines"
with their own IP address between which setting up networking is rather easy.
It is child's play to manage the state of your servers: creating, updating and
deleting your containers (or "lightweight virtual machines") is quick and easy.

During development I prefer managing containers with a lightweight solution like
`docker-compose` or bare shell scripts communicating with docker directly.

For production a more complex tool like kubernetes can solve otherwise difficult
problems for you (e.g.: network setup, load balancing, high availability).

## 3b. Running integration tests against the deployed servers

In the first steps of our workflow we have written plenty of tests. Why are we
doing it again in this stage? Why don't we write tests only at the beginning of
the workflow while developing business logic or only here as a last step?
Here are the reasons:

Why are we writing business logic tests during development in workflow step #1?

- They are much easier to write and execute than infrastructure integration tests:
    - Makes use of the go testing framework
    - No need for transport layer code, server executables, infrastructure setup
    - Allows rapid business logic prototyping
- They make it very easy to mock the business logic of complete services to
  simulate specific states/conditions (e.g.: a situation in which a service has
  inconsistent state and returns an error). Such a test in an infrastructure
  would require writing a mock server executable and the modification of the
  infrastructure before the execution of the test. It is too much work, too
  messy and text execution takes more time.

Why do we have to write infrastructure integration tests in this step?

- The tests written earlier in step #1 test only the business logic and the
  integration of the business logic residing in multiple services.
- Infrastructure integration tests ensure that everything else around the
  business logic is OK and work well together. Tested things include:
    - Transport layer code
    - Network setup
    - Other system/OS related setup (e.g.: firewalls, sysctl settings)

How to decide where to put a new test:

- A test that tests business logic should be implemented in workflow step #1
  whenever possible. These tests are much easier to write and take far less time
  to implement and execute.
- Since infrastructure integration tests are more difficult to write and execute
  I would minimise the number of these tests. This is how I define minimum:
  It should simulate an interaction between your deployed cluster and a client
  and the tests should assert that the successful interaction/flow works from
  start to finish. E.g.: An on-boarding/registration process wouldn't be blocked
  by a server side error/bug. As you see this minimum can be quite a few tests.
  Some tests that check error scenarios might also be useful: e.g.: testing your
  error reporting/monitoring systems. While implementing these tests is more
  complex, refactoring them is rarely needed since the public APIs used by them
  usually don't change very often. This is actually an advantage of these tests.

The infrastructure integration test simply performs a series of requests against
the deployed cluster. Note that the infrastructure integration tests can be
written in any language. In the simplest case you can use unix shell scripts and
command line tools. In case of this project I kept it very simple: I've written
a go client that can communicate with `svc2` that is exposed by the cluster to
the outside world. It could be used from scripts to automate infrastructure
integration testing.
 
In a more extreme case the integration tests of the backend can be merged with
those of the frontend (as an e2e test suite) but the best is to have both of
these to allow frontend and backend teams to work without friction.
If you don't have enough people then having only one frontend+backend e2e test
suite might be more practical than maintaining a lot of different kinds of tests.

It is an implementation detail but I've written the test client in go because
this way it was very easy to write a client that is compatible with the
transport layer of the server. On the other hand the structure of the test
client is pretty much the same as that of a business logic test I described in
workflow step #1 but this test client has a `main()` instead of a `TestMain()`
and accepts commandline arguments.

Here is how the test client interacts with our deployed servers:

```
+------------------------------------+     +-------------------------+
| server2a                           |     | server2b                |
|                                    |     |                         |
| +--------------------------------+ |     | +---------------------+ |
| | ServiceSet                     | |     | | ServiceSet          | |
| |                                | |     | |                     | |
| |       +-----+  +-----+         | |     | |     +-------+       | |
| |       |     |  |     |         | |     | |     |       |       | |
| |       v     |  |     v         | |     | |     |       v       | |
| |  +------+ +-+--+-+ +--------+  | |     | |  +--+---+ +------+  | |
| |  | svc1 | | svc2 | | svc3   |  | |     | |  | svc3 | | svc4 |  | |
| |  +------+ +------+ | client +--+-+--+  | |  +------+ +------+  | |
| |              ^     +--------+  | |  |  | |     ^               | |
| |              |                 | |  |  | |     |               | |
| +--------------+-----------------+ |  |  | +-----+---------------+ |
|                |                   |  |  |       |                 |
| +--------------+-----------------+ |  |  | +-----+---------------+ |
| | Listener     |                 | |  |  | |     |      Listener | |
| |           +--+---+             | |  |  | |  +--+---+           | |
| |           | svc2 |             | |  |  | |  | svc3 |           | |
| |           | cfg  |<-+          | |  +--+-+->| cfg  |           | |
| |           +------+  |          | |     | |  +------+           | |
| +---------------------+----------+ |     | +---------------------+ |
+-----------------------+------------+     +-------------------------+
                        |
+-----------------------+----+
| test_client           |    |
|                       |    |
| +---------------------+--+ |
| | ServiceSet          |  | |
| |                     |  | |
| |         +--------+  |  | |
| |         | svc2   +--+  | |
| |         | client |     | |
| |         +---+----+     | |
| |             ^          | |
| |             |          | |
| +-------------+----------+ |
|               |            |
|            +--+---+        |
|            | main |        |
|            +------+        |
+----------------------------+
```

Source code:
[server2a](examples/example1/servers/server2a/main.go)
[server2b](examples/example1/servers/server2b/main.go)
[test_client](examples/example1/servers/test_client/main.go)

# Random thoughts

## Interacting with external services, databases

It is recommended to implement the clients of external services (e.g.: twitter)
and databases as implementations of `nano.Service`. This way mocking them in
tests becomes extremely easy.

## No fancy framework features?

As you've seen this "framework" defines only a basic structure for your services.
It doesn't provide you with "fancy" features like discovery, logging, telemetry,
etc... I've created some example addons for discovery and logging but that's it.
If you want for example a kubernetes discoverer then you should implement it
yourself as an addon or use third party packages directly.

Instead of providing everything in a framework-ish interface based
Inversion-of-Control style nano encourages you to explicitly choose and use
third party golang packages for your problems.
E.g.: If you want telemetry then choose a telemetry package, initialise it at
startup time from the main of your test/server executable and use it.

## Client interface VS request-response message pairs

As you've noticed request and response objects in nano are passed with
`interface{}` type which means you can send request objects of any type using
the `Client.Request` method and services can receive them in their
`Service.Handler` without compilation errors.

This has a couple of advantages and disadvantages compared to a solution where
you write a service specific client interface with specific methods for each
request the service can perform.

Advantages:

- This solution requires much less code:
    - Only one generic `Client` implementation is needed.
    - The service handler is one method (`Service.Handle`) that uses a type
      switch to handle different request types. This makes writing mock services
      a trivial task.
- Unlike service specific client implementations the client can't contain
  service or request specific logic. This makes it very easy to implement
  clients in other languages by simply using the request/response message
  descriptions and the rules of the transport implementation in use.
  (E.g.: You can define the interface of your service with protobuf definitions
  and it will be easy to access from any language.)

Disadvantages:

- Using `interface{}` to pass request and response objects might convert some
  compile time errors into runtime errors. In practice however this isn't a huge
  problem if you write tests.
- In case of service specific client interfaces you can use the code navigation
  features of your IDE to find usages of a given interface method to find usages
  of a specific request. By using a "generic client" you have to do a global
  text search for a specific `service.Request` struct to find places where a
  specific request is made. Well, for me this isn't a huge negative.

## The error return value of `Service.Handle`

The business logic should be free of transport specific error codes (e.g.: HTTP
status codes). It is recommended to define your own error type(s) specific to
your business logic and communicate errors with those.

For practical reasons I recommend using a few error types (in best case only 1)
in your business logic. The reason for this is that there might be a transport
layer (e.g: HTTP) between your `Client.Request` and `Service.Handle` methods
and your transport implementation is responsible for transferring the error
object back to the client.
If you use only a few error types (or best: only one) then you can easily
implement the transport for those errors and transfer only the error message
of unhandled error types.

In case of the example nano addons I've implemented `NanoError` (see
[nano_error.go](addons/util/nano_error.go)) that consists of an error code
(e.g.: `"C-NOT-FOUND"`) and an error message (returned by `error.Error`).
The example http transport addon (`"github.com/pasztorpisti/nano/addons/transport/http"`)
recognises `NanoError` and can send it back to the client as it is but in case
of other error types it can transfer only the error message without the actual
error object type.

Aside from being able to serialize `NanoError` it recognises a few specific
error codes of `NanoError` and translates them to proper HTTP error status
codes. E.g.: If the error code of a `NanoError` starts with `"C-"` then
it is treated as a client error and returned with HTTP status 400, in other case
it is returned with 500. Some other specific error codes translate to specific
HTTP status codes - e.g.: `"C-NOT-FOUND"` is returned with 404.
(see: [error_codes.go](addons/transport/http/config/error_codes.go))

Note that the above error_code + error_message combo is very simple and easy to
use when it comes to cooperating with clients/servers written in other languages.
Since it is free of HTTP specific things you could easily use it with a
different transport implementation like grpc.

When you compile the services into the same executable (for example in case of
business logic tests) the previously mentioned problem doesn't exist: any error
object returned by `Service.Handle` is returned as it is by `Client.Request`
within the same process. However when the error response is transferred through
network between `Service.Handle` and `Client.Request` the transport layer
supports only a few error types (in our case only `NanoError`). This can result
in  different behavior between our tests and server executables: tests might be
able to detect certain error types that aren't supported and transferred by the
transport layer in an actual infrastructure. This can lead to successful tests
and failing error detection when the tested services are used with an actual
infrastructure. To avoid this problem we hook the `nano.NewClient` function in
our test initialisation code in order to return a `nano.Client` that simulates
the error transfer mechanism of the transport layer of our choice. The hooked
client will always return `nil` or `NanoError` during tests.
You can find test init code that hooks `nano.NewClient`
[here](examples/example1/config/test/config.go).

## Authentication & authorization

My example project doesn't implement authentication and authorization but I
think the right place to implement it is the transport layer. If the business
logic needs the user info/id of the client then the transport layer can pass
it to the business logic in the `nano.Ctx` structure.

To make the implementation more modular the listener could use a service to
delegate the authentication/authorization. Note that `Listener.Init` receives a
`ServiceSet` from which the listener can lookup a local `auth` service that has
been placed to the `ServiceSet` of the service executable for this purpose.

## Shared global variables

Since a service might be compiled into an executable (server or test) along with
other services it is better to avoid modifying shared global variables in your
service (init) code.
This is especially relevant in case of global variables declared by the standard
library (e.g.: `http.DefaultClient`). These shared global variables should have
well defined values set up at startup time by the executable (in `main`,
or `TestMain`).

Your services can still have their own non-shared global variables but to make
your code robust it is recommended to keep state inside your service objects
instead of global scope.

## Test executables

Note that go compiles all test files of a package into a single test executable.
you can compile separate test executables from each `.go` file of the same
package by explicitly specifying the go file(s) for the `go test` command
instead of a package spec (similarly how you can do this with `go run`) but this
seems to be an undocumented feature so probably best to avoid. The go tool
encourages thinking in packages, not source files.

Why am I mentioning this? If your services make use of global variables then you
might not be able to create and initialise a service multiple times inside one
executable during one execution. This means that it is a recommended pattern to
create and initialise only one `nano.ServiceSet` in each test package (which
compiles into a single executable) in `TestMain` and every test function in the
package should use the `nano.ServiceSet` that has been initialised by `TestMain`
at the startup time of the test executable.

This is why [test1](examples/example1/services/svc2/tests/test1/svc_test.go) and
[test2](examples/example1/services/svc2/tests/test2/svc_test.go) have been
implemented in their own packages.
 
By using the undocumented `go test <go_file(s)>` command we could put all tests
into one directory but in that case it would be recommended to remove the
`_test` suffix from the filenames to prevent the automatic `go` test discovery
from recognising them as test files and compile them as a package into a single
executable.

## Graceful server shutdown

A server should respond to `SIGINT` and `SIGTERM` signals gracefully. The
listeners should stop accepting new connections and waiting for any outstanding
requests that are being served at the time of receiving the signal.

This hasn't been implemented by the listener of the nano http transport but
demonstrating this is ot of the scope of this project.

Currently the standard `net/http` package doesn't directly support graceful
shutdown but you can find tutorials on the net on how to hack graceful shutdown
into your http server and there are plans to add support for it to the standard
`net/http` package.
