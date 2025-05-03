package mock

import (
	"context"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxpvclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv/client"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

var _ azurerwxpvclient.Client = &fileShareStore{}

func newFileShareStore(subscription string) *fileShareStore {
	return &fileShareStore{
		subscription: subscription,
		fileshares:   make(map[string]*armstorage.FileShareItem),
	}
}

type fileShareStore struct {
	m            sync.Mutex
	subscription string
	fileshares   map[string]*armstorage.FileShareItem
}

func (s *fileShareStore) CreateFileShare(ctx context.Context, id string) error {
	s.m.Lock()
	defer s.m.Unlock()

	_, _, fileShareName, _, _, err := client.ParsePvVolumeHandle(id)
	if err != nil {
		return err
	}

	//Add to the map.
	s.fileshares[id] = &armstorage.FileShareItem{
		Name: to.Ptr(fileShareName),
		ID:   to.Ptr(id),
	}

	composed.LoggerFromCtx(ctx).Info("mock: Create/Update FileShare", "share-id", id, "count", len(s.fileshares))

	return nil
}

func (s *fileShareStore) GetFileShare(ctx context.Context, id string) (*armstorage.FileShareItem, error) {
	s.m.Lock()
	defer s.m.Unlock()

	fs := s.fileshares[id]
	composed.LoggerFromCtx(ctx).Info("mock: GetFileShare()", "length", len(s.fileshares))

	return fs, nil
}

func (s *fileShareStore) DeleteFileShare(ctx context.Context, id string) error {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.fileshares, id)
	composed.LoggerFromCtx(ctx).Info("mock: DeleteFileShare()", "id", id, "length", len(s.fileshares))
	return nil
}
