// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package ntfmock

import (
	context "context"
	"github.com/anvh2/trading-bot/pkg/api/v1/notifier"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	sync "sync"
)

// Ensure, that NotifierServiceClientMock does implement notifier.NotifierServiceClient.
// If this is not the case, regenerate this file with moq.
var _ notifier.NotifierServiceClient = &NotifierServiceClientMock{}

// NotifierServiceClientMock is a mock implementation of notifier.NotifierServiceClient.
//
// 	func TestSomethingThatUsesNotifierServiceClient(t *testing.T) {
//
// 		// make and configure a mocked notifier.NotifierServiceClient
// 		mockedNotifierServiceClient := &NotifierServiceClientMock{
// 			PushFunc: func(ctx context.Context, in *notifier.PushRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
// 				panic("mock out the Push method")
// 			},
// 		}
//
// 		// use mockedNotifierServiceClient in code that requires notifier.NotifierServiceClient
// 		// and then make assertions.
//
// 	}
type NotifierServiceClientMock struct {
	// PushFunc mocks the Push method.
	PushFunc func(ctx context.Context, in *notifier.PushRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)

	// calls tracks calls to the methods.
	calls struct {
		// Push holds details about calls to the Push method.
		Push []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// In is the in argument value.
			In *notifier.PushRequest
			// Opts is the opts argument value.
			Opts []grpc.CallOption
		}
	}
	lockPush sync.RWMutex
}

// Push calls PushFunc.
func (mock *NotifierServiceClientMock) Push(ctx context.Context, in *notifier.PushRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if mock.PushFunc == nil {
		panic("NotifierServiceClientMock.PushFunc: method is nil but NotifierServiceClient.Push was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		In   *notifier.PushRequest
		Opts []grpc.CallOption
	}{
		Ctx:  ctx,
		In:   in,
		Opts: opts,
	}
	mock.lockPush.Lock()
	mock.calls.Push = append(mock.calls.Push, callInfo)
	mock.lockPush.Unlock()
	return mock.PushFunc(ctx, in, opts...)
}

// PushCalls gets all the calls that were made to Push.
// Check the length with:
//     len(mockedNotifierServiceClient.PushCalls())
func (mock *NotifierServiceClientMock) PushCalls() []struct {
	Ctx  context.Context
	In   *notifier.PushRequest
	Opts []grpc.CallOption
} {
	var calls []struct {
		Ctx  context.Context
		In   *notifier.PushRequest
		Opts []grpc.CallOption
	}
	mock.lockPush.RLock()
	calls = mock.calls.Push
	mock.lockPush.RUnlock()
	return calls
}
