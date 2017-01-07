package nano

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// testSvc implements the Service interface.
type testSvc struct {
	name    string
	handler func(c *Ctx, req interface{}) (interface{}, error)
}

func (p *testSvc) Name() string {
	return p.name
}

func (p *testSvc) Handle(c *Ctx, req interface{}) (interface{}, error) {
	return p.handler(c, req)
}

// testSvcInit implements the ServiceInit interface.
type testSvcInit func(cs ClientSet) error

func (f testSvcInit) Init(cs ClientSet) error {
	return f(cs)
}

// testSvcInitFinished implements the ServiceInitFinished interface.
type testSvcInitFinished func() error

func (f testSvcInitFinished) InitFinished() error {
	return f()
}

// testListener implements the Listener interface.
type testListener struct {
	init   func(ServiceSet) error
	listen func() error
}

func (p *testListener) Init(ss ServiceSet) error {
	if p.init == nil {
		return nil
	}
	return p.init(ss)
}

func (p *testListener) Listen() error {
	if p.listen == nil {
		return nil
	}
	return p.listen()
}

// eventSetEquals check if the events and want arrays contain the same set of
// strings ignoring the ordering.
func eventSetEquals(t *testing.T, events []string, want ...string) bool {
	if len(events) != len(want) {
		t.Errorf("event sets differ: events=%v, want=%v", events, want)
		return false
	}
	eventSet := make(map[string]struct{}, len(events))
	for _, e := range events {
		eventSet[e] = struct{}{}
	}
	for _, e := range want {
		if _, ok := eventSet[e]; !ok {
			t.Errorf("event sets differ: events=%v, want=%v", events, want)
			return false
		}
	}
	return true
}

// eventSetContains check if the specified subset is contained by the events array.
func eventSetContains(t *testing.T, events []string, subset ...string) bool {
	if len(events) < len(subset) {
		t.Errorf("event set doesn't contain subset: events=%v, subset=%v", events, subset)
		return false
	}
	eventSet := make(map[string]struct{}, len(events))
	for _, e := range events {
		eventSet[e] = struct{}{}
	}
	for _, e := range subset {
		if _, ok := eventSet[e]; !ok {
			t.Errorf("event set doesn't contain subset: events=%v, subset=%v", events, subset)
			return false
		}
	}
	return true
}

func TestNewServiceSet(t *testing.T) {
	var events []string
	evt := func(e string) {
		events = append(events, e)
	}

	handler := func(c *Ctx, req interface{}) (interface{}, error) {
		evt("Handler:" + c.Svc.Name())
		return nil, nil
	}

	svc1 := &testSvc{
		name:    "svc1",
		handler: handler,
	}

	svc2 := &struct {
		testSvc
		testSvcInit
	}{
		testSvc: testSvc{
			name:    "svc2",
			handler: handler,
		},
		testSvcInit: func(cs ClientSet) error {
			evt("Init:svc2")
			return nil
		},
	}

	svc3 := &struct {
		testSvc
		testSvcInitFinished
	}{
		testSvc: testSvc{
			name:    "svc3",
			handler: handler,
		},
		testSvcInitFinished: func() error {
			evt("InitFinished:svc3")
			return nil
		},
	}

	svc4 := &struct {
		testSvc
		testSvcInit
		testSvcInitFinished
	}{
		testSvc: testSvc{
			name:    "svc4",
			handler: handler,
		},
		testSvcInit: func(cs ClientSet) error {
			evt("Init:svc4")
			return nil
		},
		testSvcInitFinished: func() error {
			evt("InitFinished:svc4")
			return nil
		},
	}

	// create and initialise the service set
	_ = NewServiceSet(svc2, svc4, svc1, svc3)

	if v, want := len(events), 4; v != want {
		t.Errorf("number of events is %v, want %v - events: %v", v, want, events)
	}
	eventSetEquals(t, events[0:2], "Init:svc2", "Init:svc4")
	eventSetEquals(t, events[2:4], "InitFinished:svc3", "InitFinished:svc4")
}

func TestNewServiceSet_Initialisation_Calls_NewClientSet(t *testing.T) {
	var events []string
	origNewClientSet := NewClientSet
	defer func() {
		NewClientSet = origNewClientSet
	}()
	NewClientSet = func(ss ServiceSet, ownerName string) ClientSet {
		events = append(events, ownerName)
		return origNewClientSet(ss, ownerName)
	}

	handler := func(c *Ctx, req interface{}) (interface{}, error) {
		return nil, nil
	}
	init := func(cs ClientSet) error {
		return nil
	}

	svc1 := &struct {
		testSvc
		testSvcInit
	}{
		testSvc: testSvc{
			name:    "svc1",
			handler: handler,
		},
		testSvcInit: init,
	}

	svc2 := &struct {
		testSvc
		testSvcInit
	}{
		testSvc: testSvc{
			name:    "svc2",
			handler: handler,
		},
		testSvcInit: init,
	}

	_ = NewServiceSet(svc2, svc1)

	// Calling NewServiceSet should involve at least two calls to NewClientSet
	// for the svc1 and svc2 services.
	eventSetContains(t, events, "svc1", "svc2")
}

func TestClientSet_LookupClient_Calls_NewClient(t *testing.T) {
	var svcs []Service
	var owners []string
	origNewClient := NewClient
	defer func() {
		NewClient = origNewClient
	}()
	NewClient = func(svc Service, ownerName string) Client {
		svcs = append(svcs, svc)
		owners = append(owners, ownerName)
		return origNewClient(svc, ownerName)
	}

	svc1 := &testSvc{
		name: "svc1",
	}
	ss := NewServiceSet(svc1)
	cs := NewClientSet(ss, "test")

	lenBeforeLookup := len(svcs)
	_ = cs.LookupClient("svc1")

	found := false
	for i := lenBeforeLookup; i < len(svcs); i++ {
		if svcs[i] == svc1 && owners[i] == "test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("ClientSet.LookupClient didn't call NewClient")
	}
}

func TestServiceSet_LookupService(t *testing.T) {
	svc1 := &testSvc{
		name: "svc1",
	}
	svc2 := &testSvc{
		name: "svc2",
	}

	ss := NewServiceSet(svc1, svc2)

	_, err := ss.LookupService("svc1")
	if err != nil {
		t.Errorf("error looking up svc1 :: %v", err)
	}

	_, err = ss.LookupService("svc2")
	if err != nil {
		t.Errorf("error looking up svc2 :: %v", err)
	}

	_, err = ss.LookupService("svc3")
	if err == nil {
		t.Error("unexpected success while looking up svc3")
	}
}

func TestClient_Request_Nil_Ctx(t *testing.T) {
	const ownerName = "test"

	called := false
	var ctx *Ctx

	svc1 := &testSvc{
		name: "svc1",
		handler: func(c *Ctx, req interface{}) (interface{}, error) {
			called = true
			ctx = c
			return nil, nil
		},
	}

	ss := NewServiceSet(svc1)
	cs := NewClientSet(ss, ownerName)
	client := cs.LookupClient("svc1")
	_, _ = client.Request(nil, nil)

	if !called {
		t.Error("client wasn't called")
		t.FailNow()
	}
	if ctx == nil {
		t.Error("ctx is nil")
		t.FailNow()
	}
	if len(ctx.ReqID) != GeneratedReqIDBytesLen*2 {
		t.Errorf("len(ctx.ReqID) == %v, want %v", len(ctx.ReqID), GeneratedReqIDBytesLen*2)
	}
	if ctx.Context == nil {
		t.Error("ctx.Context is nil")
	}
	if ctx.Svc != svc1 {
		t.Error("ctx.Svc isn't set correctly")
	}
	if ctx.ClientName != ownerName {
		t.Errorf("ctx.ClientName == %q, want %q", ctx.ClientName, ownerName)
	}
}

func TestClient_Request_Empty_Ctx(t *testing.T) {
	const ownerName = "test"

	called := false
	var ctx *Ctx

	svc1 := &testSvc{
		name: "svc1",
		handler: func(c *Ctx, req interface{}) (interface{}, error) {
			called = true
			ctx = c
			return nil, nil
		},
	}

	ss := NewServiceSet(svc1)
	cs := NewClientSet(ss, ownerName)
	client := cs.LookupClient("svc1")
	_, _ = client.Request(&Ctx{}, nil)

	if !called {
		t.Error("client wasn't called")
		t.FailNow()
	}
	if ctx == nil {
		t.Error("ctx is nil")
		t.FailNow()
	}
	if len(ctx.ReqID) != GeneratedReqIDBytesLen*2 {
		t.Errorf("len(ctx.ReqID) == %v, want %v", len(ctx.ReqID), GeneratedReqIDBytesLen*2)
	}
	if ctx.Context == nil {
		t.Error("ctx.Context is nil")
	}
	if ctx.Svc != svc1 {
		t.Error("ctx.Svc isn't set correctly")
	}
	if ctx.ClientName != ownerName {
		t.Errorf("ctx.ClientName == %q, want %q", ctx.ClientName, ownerName)
	}
}

func TestClient_Request_With_Ctx(t *testing.T) {
	const ownerName = "test"

	called := false
	var ctx *Ctx

	svc1 := &testSvc{
		name: "svc1",
		handler: func(c *Ctx, req interface{}) (interface{}, error) {
			called = true
			ctx = c
			return nil, nil
		},
	}

	ss := NewServiceSet(svc1)
	cs := NewClientSet(ss, ownerName)
	client := cs.LookupClient("svc1")

	myContext := context.Background()
	const reqID = "MyReqID"
	inputCtx := &Ctx{
		ReqID:      reqID,
		Context:    myContext,
		Svc:        &testSvc{name: "ignored"},
		ClientName: "ignored",
	}
	_, _ = client.Request(inputCtx, nil)

	if !called {
		t.Error("client wasn't called")
		t.FailNow()
	}
	if ctx == nil {
		t.Error("ctx is nil")
		t.FailNow()
	}
	if ctx.ReqID != reqID {
		t.Errorf("ctx.ReqID == %q, want %q", ctx.ReqID, reqID)
	}
	if ctx.Context == nil {
		t.Error("ctx.Context is nil")
	} else if ctx.Context == myContext {
		t.Error("ctx.Context shouldn't be the same as myContext")
	}
	if ctx.Svc != svc1 {
		t.Error("ctx.Svc isn't set correctly")
	}
	if ctx.ClientName != ownerName {
		t.Errorf("ctx.ClientName == %q, want %q", ctx.ClientName, ownerName)
	}
}

func testNewReqID(t *testing.T, generatedReqIDBytesLen int) {
	origLen := GeneratedReqIDBytesLen
	defer func() {
		GeneratedReqIDBytesLen = origLen
	}()
	GeneratedReqIDBytesLen = generatedReqIDBytesLen

	reqID, err := NewReqID()
	if err != nil {
		t.Errorf("NewReqID failed :: %v", err)
		t.FailNow()
	}
	if len(reqID) != GeneratedReqIDBytesLen*2 {
		t.Errorf("len(reqID) == %v, want %v", len(reqID), GeneratedReqIDBytesLen*2)
	}
}

func TestNewReqID(t *testing.T) {
	testNewReqID(t, GeneratedReqIDBytesLen)
}

func TestNewReqID_With_Custom_GeneratedReqIDBytesLen(t *testing.T) {
	testNewReqID(t, 5)
}

func isContextCanceled(c context.Context) bool {
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}

const waitAttempts = 100
const waitDuration = time.Millisecond * 1

func waitForCancel(t *testing.T, c context.Context) bool {
	for i := 0; i < waitAttempts; i++ {
		if isContextCanceled(c) {
			return true
		}
		time.Sleep(waitDuration)
	}
	return false
}

func TestNewContext_Nil_Parent(t *testing.T) {
	c, origCancel := NewContext(nil)
	cancel := func() {
		if origCancel != nil {
			origCancel()
			origCancel = nil
		}
	}
	defer cancel()

	if isContextCanceled(c) {
		t.Error("context is canceled right after creation")
	}

	cancel()
	if !waitForCancel(t, c) {
		t.Error("context isn't canceled after canceling it")
	}
}

func TestNewContext_With_Parent(t *testing.T) {
	parentContext, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	c, origCancel := NewContext(parentContext)
	cancel := func() {
		if origCancel != nil {
			origCancel()
			origCancel = nil
		}
	}
	defer cancel()

	if isContextCanceled(c) {
		t.Error("context is canceled right after creation")
	}

	cancel()
	if !waitForCancel(t, c) {
		t.Error("context isn't canceled after canceling it")
	}
	if isContextCanceled(parentContext) {
		t.Error("parent is canceled")
	}
}

func TestNewContext_With_Parent_Cancel(t *testing.T) {
	parentContext, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	c, cancel := NewContext(parentContext)
	defer cancel()

	if isContextCanceled(c) {
		t.Error("context is canceled right after creation")
	}

	parentCancel()
	if !waitForCancel(t, c) {
		t.Error("context isn't canceled after canceling it")
	}
	if !waitForCancel(t, parentContext) {
		t.Error("parent isn't canceled")
	}
}

func TestRunServer(t *testing.T) {
	var events []string
	var mu sync.Mutex
	evt := func(e string) {
		// Synchronization is needed because the Listen() method of Listeners
		// is called on separate goroutines.
		mu.Lock()
		defer mu.Unlock()
		events = append(events, e)
	}

	onError := func(l Listener, err error) {
		t.Errorf("listener %p failed :: %v", l, err)
	}
	ss := NewServiceSet(&testSvc{
		name: "svc1",
	})

	listener1 := &testListener{
		init: func(s ServiceSet) error {
			if _, err := s.LookupService("svc1"); err != nil {
				t.Error("error looking up svc1")
			}
			evt("Init:listener1")
			return nil
		},
		listen: func() error {
			evt("Listen:listener1")
			return nil
		},
	}
	listener2 := &testListener{
		init: func(s ServiceSet) error {
			if _, err := s.LookupService("svc1"); err != nil {
				t.Error("error looking up svc1")
			}
			evt("Init:listener2")
			return nil
		},
		listen: func() error {
			evt("Listen:listener2")
			return nil
		},
	}

	runServer(onError, ss, listener2, listener1)

	if v, want := len(events), 4; v != want {
		t.Errorf("number of events is %v, want %v - events: %v", v, want, events)
	}
	eventSetEquals(t, events[0:2], "Init:listener1", "Init:listener2")
	eventSetEquals(t, events[2:4], "Listen:listener1", "Listen:listener2")
}

func TestRunServer_Listener_Init_Error(t *testing.T) {
	onError := func(l Listener, err error) {
		t.Errorf("listener %p failed :: %v", l, err)
	}
	ss := NewServiceSet(&testSvc{
		name: "svc1",
	})
	e := errors.New("listener init error")
	called := false
	listener := &testListener{
		init: func(s ServiceSet) error {
			called = true
			return e
		},
	}

	defer func() {
		if !called {
			t.Error("listener init wasn't called")
		}
		v := recover()
		if v == nil {
			t.Error("haven't received the expected panic")
		} else {
			s := fmt.Sprintf("%v", v)
			if !strings.Contains(s, e.Error()) {
				t.Error("panic message doesn't contain the error message")
			}
		}
	}()

	runServer(onError, ss, listener)
}

func TestRunServer_Listen_Error(t *testing.T) {
	e := errors.New("listen error")
	called := false
	onError := func(l Listener, err error) {
		called = true
		if err != e {
			t.Errorf("listener %p failed with unexpected error :: %v", l, err)
		}
	}

	ss := NewServiceSet(&testSvc{
		name: "svc1",
	})
	listener := &testListener{
		listen: func() error {
			return e
		},
	}

	runServer(onError, ss, listener)

	if !called {
		t.Error("onError wasn't called")
	}
}
