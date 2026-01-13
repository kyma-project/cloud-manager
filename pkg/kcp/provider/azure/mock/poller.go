package mock

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	azruntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

var _ azureclient.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse] = (*Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse])(nil)

func NewPollerMock[T any](resp T, err error, resumeToken string) *Poller[T] {
	return &Poller[T]{
		resp:        resp,
		err:         err,
		resumeToken: resumeToken,
	}
}

type Poller[T any] struct {
	resp        T
	err         error
	resumeToken string
}

func (p *Poller[T]) SetResponse(resp T) {
	p.resp = resp
}

func (p *Poller[T]) SetError(err error) {
	p.err = err
}

func (p *Poller[T]) SetResumeToken(resumeToken string) {
	p.resumeToken = resumeToken
}

// implement azureclient.Poller =============================================================

func (p *Poller[T]) PollUntilDone(ctx context.Context, options *azruntime.PollUntilDoneOptions) (resp T, err error) {
	time.Sleep(5 * time.Millisecond)
	return p.resp, p.err
}

func (p *Poller[T]) Poll(ctx context.Context) (resp *http.Response, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *Poller[T]) Done() bool {
	return !reflect.ValueOf(p.resp).IsNil()
}

func (p *Poller[T]) Result(ctx context.Context) (res T, err error) {
	if !p.Done() {
		err = errors.New("poller is in a non-terminal state")
		return
	}
	return p.resp, p.err
}

func (p *Poller[T]) ResumeToken() (string, error) {
	if p.Done() {
		return "", errors.New("poller is in a terminal state")
	}
	return p.resumeToken, p.err
}
