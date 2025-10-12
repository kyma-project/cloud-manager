package fake

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

var _ cluster.Cluster = &Cluster{}

type Cluster struct {
	StartCallCount int
	ErrPre         error
	ErrPost        error

	HttpClient    *http.Client
	RestConfig    *rest.Config
	Cache         *Cache
	Scheme        *runtime.Scheme
	Client        client.Client
	FieldIndexer  client.FieldIndexer
	EventRecorder record.EventRecorder
	RESTMapper    meta.RESTMapper
	APIReader     client.Reader
}

func (r *Cluster) Start(ctx context.Context) error {
	r.StartCallCount++
	if r.ErrPre != nil {
		return r.ErrPre
	}
	<-ctx.Done()
	return r.ErrPost
}

func (f *Cluster) GetHTTPClient() *http.Client {
	return f.HttpClient
}

func (f *Cluster) GetConfig() *rest.Config {
	return f.RestConfig
}

func (f *Cluster) GetCache() cache.Cache {
	return f.Cache
}

func (f *Cluster) GetScheme() *runtime.Scheme {
	return f.Scheme
}

func (f *Cluster) GetClient() client.Client {
	return f.Client
}

func (f *Cluster) GetFieldIndexer() client.FieldIndexer {
	return f.FieldIndexer
}

func (f *Cluster) GetEventRecorderFor(name string) record.EventRecorder {
	return f.EventRecorder
}

func (f *Cluster) GetRESTMapper() meta.RESTMapper {
	return f.RESTMapper
}

func (f *Cluster) GetAPIReader() client.Reader {
	return f.APIReader
}
