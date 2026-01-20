package client

import (
	"context"
	"net/http"
	"time"

	azruntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

type Poller[T any] interface {
	// PollUntilDone will poll the service endpoint until a terminal state is reached, an error is received, or the context expires.
	// It internally uses Poll(), Done(), and Result() in its polling loop, sleeping for the specified duration between intervals.
	// options: pass nil to accept the default values.
	// NOTE: the default polling frequency is 30 seconds which works well for most operations.  However, some operations might
	// benefit from a shorter or longer duration.
	PollUntilDone(ctx context.Context, options *azruntime.PollUntilDoneOptions) (resp T, err error)

	// Poll fetches the latest state of the LRO.  It returns an HTTP response or error.
	// If Poll succeeds, the poller's state is updated and the HTTP response is returned.
	// If Poll fails, the poller's state is unmodified and the error is returned.
	// Calling Poll on an LRO that has reached a terminal state will return the last HTTP response.
	Poll(ctx context.Context) (resp *http.Response, err error)

	// Done returns true if the LRO has reached a terminal state.
	// Once a terminal state is reached, call Result().
	Done() bool

	// Result returns the result of the LRO and is meant to be used in conjunction with Poll and Done.
	// If the LRO completed successfully, a populated instance of T is returned.
	// If the LRO failed or was canceled, an *azcore.ResponseError error is returned.
	// Calling this on an LRO in a non-terminal state will return an error.
	Result(ctx context.Context) (res T, err error)

	// ResumeToken returns a value representing the poller that can be used to resume
	// the LRO at a later time. ResumeTokens are unique per service operation.
	// The token's format should be considered opaque and is subject to change.
	// Calling this on an LRO in a terminal state will return an error.
	ResumeToken() (string, error)
}

func PollUntilDone[T any](poller Poller[T], err error) func(ctx context.Context, options *azruntime.PollUntilDoneOptions) (res T, err error) {
	return func(ctx context.Context, options *azruntime.PollUntilDoneOptions) (res T, err error) {
		if err != nil {
			var zero T
			return zero, err
		}
		if options == nil {
			options = &azruntime.PollUntilDoneOptions{
				Frequency: 5 * time.Second,
			}
		}
		return poller.PollUntilDone(ctx, options)
	}
}
