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

var _ azureclient.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse] = (*PollerMock[armnetwork.VirtualNetworksClientCreateOrUpdateResponse])(nil)

func NewPollerMock[T any](resp T, err error, resumeToken string) *PollerMock[T] {
	return &PollerMock[T]{
		resp:        resp,
		err:         err,
		resumeToken: resumeToken,
	}
}

type PollerMock[T any] struct {
	resp        T
	err         error
	resumeToken string
}

func (p *PollerMock[T]) SetResponse(resp T) {
	p.resp = resp
}

func (p *PollerMock[T]) SetError(err error) {
	p.err = err
}

func (p *PollerMock[T]) SetResumeToken(resumeToken string) {
	p.resumeToken = resumeToken
}

// implement azureclient.Poller =============================================================

func (p *PollerMock[T]) PollUntilDone(ctx context.Context, options *azruntime.PollUntilDoneOptions) (resp T, err error) {
	time.Sleep(5 * time.Millisecond)
	return p.resp, p.err
}

func (p *PollerMock[T]) Poll(ctx context.Context) (resp *http.Response, err error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *PollerMock[T]) Done() bool {
	return !reflect.ValueOf(p.resp).IsNil()
}

func (p *PollerMock[T]) Result(ctx context.Context) (res T, err error) {
	if !p.Done() {
		err = errors.New("poller is in a non-terminal state")
		return
	}
	return p.resp, p.err
}

func (p *PollerMock[T]) ResumeToken() (string, error) {
	if p.Done() {
		return "", errors.New("poller is in a terminal state")
	}
	return p.resumeToken, p.err
}
