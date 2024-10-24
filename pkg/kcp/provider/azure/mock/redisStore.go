package mock

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

var _ RedisInstanceClient = &redisStore{}
var _ RedisConfig = &redisStore{}

func newRedisStore(subscription string) *redisStore {
	return &redisStore{
		subscription: subscription,
		items:        map[string]map[string]*instanceInfo{},
	}
}

type instanceInfo struct {
	redis      *armredis.ResourceInfo
	accessKeys *armredis.AccessKeys
}

type redisStore struct {
	m sync.Mutex

	subscription string

	// items is a map of resourceGroup => redisName => *armredis.ResourceInfo
	items map[string]map[string]*instanceInfo
}

// Config =================================================================================================

func (s *redisStore) AzureRemoveRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, err := s.getRedisInfoNonLocking(resourceGroupName, redisInstanceName)
	if err != nil {
		return err
	}
	delete(s.items[resourceGroupName], redisInstanceName)

	return nil
}

func (s *redisStore) AzureSetRedisInstanceState(ctx context.Context, resourceGroupName, redisInstanceName string, state armredis.ProvisioningState) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	info, err := s.getRedisInfoNonLocking(resourceGroupName, redisInstanceName)
	if err != nil {
		return err
	}

	info.redis.Properties.ProvisioningState = ptr.To(state)

	return nil
}

// RedisInstanceClient ====================================================================================

func (s *redisStore) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.items[resourceGroupName]
	if !ok {
		s.items[resourceGroupName] = map[string]*instanceInfo{}
	}
	_, ok = s.items[resourceGroupName][redisInstanceName]
	if ok {
		return fmt.Errorf("redis instance %s already exist", azureutil.NewRedisInstanceResourceId(s.subscription, resourceGroupName, redisInstanceName).String())
	}

	if parameters.Properties == nil {
		parameters.Properties = &armredis.CreateProperties{}
	}

	props := &armredis.Properties{}
	err := util.JsonCloneInto(parameters.Properties, props)
	if err != nil {
		return err
	}

	props.ProvisioningState = ptr.To(armredis.ProvisioningStateSucceeded)
	props.HostName = ptr.To("redis.tcp")
	props.SSLPort = ptr.To(int32(6380))
	if props.SKU == nil {
		props.SKU = &armredis.SKU{
			Capacity: ptr.To(int32(1)),
		}
	}

	item := &instanceInfo{
		redis: &armredis.ResourceInfo{
			Location:   parameters.Location,
			Name:       ptr.To(redisInstanceName),
			Properties: props,
			Identity:   parameters.Identity,
			Tags:       parameters.Tags,
			Zones:      parameters.Zones,
		},
		accessKeys: &armredis.AccessKeys{
			PrimaryKey:   ptr.To(uuid.NewString()),
			SecondaryKey: ptr.To(uuid.NewString()),
		},
	}

	s.items[resourceGroupName][redisInstanceName] = item

	return nil
}

func (s *redisStore) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	info, err := s.getRedisInfoNonLocking(resourceGroupName, redisInstanceName)
	if err != nil {
		return err
	}

	if parameters.Properties == nil {
		parameters.Properties = &armredis.UpdateProperties{}
	}
	props := &armredis.Properties{}
	err = util.JsonCloneInto(parameters.Properties, props)
	if err != nil {
		return err
	}

	if err = mergo.Merge(info.redis.Properties, props); err != nil {
		return err
	}

	if parameters.Identity != nil {
		if info.redis.Identity == nil {
			info.redis.Identity = parameters.Identity
		} else {
			err = mergo.Merge(info.redis.Identity, parameters.Identity)
			if err != nil {
				return err
			}
		}
	}
	if parameters.Tags != nil {
		for k, v := range parameters.Tags {
			info.redis.Tags[k] = v
		}
	}

	return nil
}

func (s *redisStore) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	info, err := s.getRedisInfoNonLocking(resourceGroupName, redisInstanceName)
	if err != nil {
		return nil, err
	}

	res, err := util.JsonClone(info.redis)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *redisStore) getRedisInfoNonLocking(resourceGroupName, redisInstanceName string) (*instanceInfo, error) {
	group, ok := s.items[resourceGroupName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	info, ok := group[redisInstanceName]
	if !ok {
		return nil, azuremeta.NewAzureNotFoundError()
	}
	return info, nil
}

func (s *redisStore) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	info, err := s.getRedisInfoNonLocking(resourceGroupName, redisInstanceName)
	if err != nil {
		return err
	}

	info.redis.Properties.ProvisioningState = ptr.To(armredis.ProvisioningStateDeleting)

	return nil
}

func (s *redisStore) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	info, err := s.getRedisInfoNonLocking(resourceGroupName, redisInstanceName)
	if err != nil {
		return nil, err
	}

	return []string{ptr.Deref(info.accessKeys.PrimaryKey, ""), ptr.Deref(info.accessKeys.SecondaryKey, "")}, nil
}
