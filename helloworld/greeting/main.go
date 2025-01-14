package main

import (
	"context"

	"github.com/iansmith/parigot-example/hello-world/g/greeting/v1"
	"github.com/iansmith/parigot/api/shared/id"
	pcontext "github.com/iansmith/parigot/context"
	"github.com/iansmith/parigot/g/syscall/v1"
	lib "github.com/iansmith/parigot/lib/go"
	"github.com/iansmith/parigot/lib/go/future"
)

// All services have a main.  The services should not really "start", however
// until Ready() is called because their dependencies are only guaranteed
// to be up when Ready() is reached.
func main() {
	// create a container for logging into
	ctx := pcontext.NewContextWithContainer(pcontext.GuestContext(context.Background()), "[greeting]main")

	// the implementation of the service (no state right now)
	impl := &myService{}

	// Init initiaizes a service and normally receives a list of functions
	// that indicate dependencies, but we don't have any here.
	binding := greeting.Init(ctx, []lib.MustRequireFunc{}, impl)

	// Run waits for calls to our single method and should not return.
	kerr := greeting.Run(ctx, binding, greeting.TimeoutInMillis, nil)

	// Should not happen.
	pcontext.Errorf(ctx, "error caused run to exit in greeting: %s", syscall.KernelErr_name[int32(kerr)])
}

// myService is the true implementation of the greeting service.
type myService struct{}

// test at compile time that myService has appropriate methods.
var _ = greeting.Greeting(&myService{})

// the values by the language number
var resultByLang = map[int32]string{
	1: "hello",
	2: "bonjour",
	3: "guten tag",
}

// fetchGreeting (with a lowercase f) is here because it is easier
// to unit test the service with this structure.  The "real" FetchGreeting
// just calls this one and deals with the returning of futures
// which are required for the real service.
func (m *myService) fetchGreeting(ctx context.Context, req *greeting.FetchGreetingRequest) (*greeting.FetchGreetingResponse, greeting.GreetErr) {
	max := len(greeting.Tongue_value) - 1 // -1 because it has a zero in it

	// protoc generates 32 bit ints for every enum value
	if req.GetTongue() < 1 || int(req.GetTongue()) > max {
		return nil, greeting.GreetErr_UnknownLang
	}
	resp := &greeting.FetchGreetingResponse{}
	resp.Greeting = resultByLang[int32(req.GetTongue())]
	return resp, greeting.GreetErr_NoError
}

// FetchGreeting returns a string that is a greeting for the
// given Tongue in the request. The future returned is already
// completed because there is no need to wait for any
// result.
func (m *myService) FetchGreeting(ctx context.Context, req *greeting.FetchGreetingRequest) *greeting.FutureFetchGreeting {
	resp, err := m.fetchGreeting(ctx, req)
	fut := greeting.NewFutureFetchGreeting()

	if err != greeting.GreetErr_NoError {
		fut.Method.CompleteMethod(ctx, nil, err)
	} else {
		// err is NoError
		fut.Method.CompleteMethod(ctx, resp, err)
	}
	return fut
}

// Ready simply returns an already completed future with the value
// false because it does not have anything to do.  Many Ready()
// functions use this function to MustLocateXXX() calls to obtain
// references to other services.  The second parameter is
// passed here with the ServiceId of myService (the receiver
// of this method call) but it is not needed.
func (m *myService) Ready(_ context.Context, _ id.ServiceId) *future.Base[bool] {
	fut := future.NewBase[bool]()
	fut.Set(true)
	return fut
}
