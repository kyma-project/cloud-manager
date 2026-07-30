package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	redisinstance "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/redisinstance/client"
)

// RedisInstanceEntry is the stored representation of a standard (non-sharded)
// AliCloud r-kvstore instance.
type RedisInstanceEntry struct {
	InstanceId       string
	InstanceName     string
	InstanceClass    string
	InstanceStatus   string
	EngineVersion    string
	VpcId            string
	VSwitchId        string
	NetworkType      string
	ChargeType       string
	Port             int64
	ConnectionDomain string
	ReadOnlyCount    int32
	Password         string
	Config           string
	SecurityIps      string
	SslEnabled       bool
	// PendingSslEnabled holds the SSL value being applied asynchronously while
	// InstanceStatus is SSLModifying. TransitionAllToNormal applies it.
	PendingSslEnabled *bool
}

// RedisClusterEntry is the stored representation of a sharded AliCloud
// r-kvstore cloud-native cluster.
type RedisClusterEntry struct {
	InstanceId       string
	InstanceName     string
	InstanceClass    string
	InstanceStatus   string
	EngineVersion    string
	VpcId            string
	VSwitchId        string
	NetworkType      string
	ChargeType       string
	Port             int64
	ConnectionDomain string
	ShardCount       int32
	ReplicasPerShard int32
	Password         string
	Config           string
	SecurityIps      string
	SslEnabled       bool
	// PendingSslEnabled holds the SSL value being applied asynchronously while
	// InstanceStatus is SSLModifying. TransitionAllToNormal applies it.
	PendingSslEnabled *bool
}

// redisStore is the shared in-memory backing store for the r-kvstore mock. It
// serves both the standard-instance and cluster views because both types use
// the same underlying AliCloud API surface.
type redisStore struct {
	m sync.Mutex

	instances map[string]*RedisInstanceEntry
	clusters  map[string]*RedisClusterEntry

	instanceErrors map[string]error
	clusterErrors  map[string]error
}

func newRedisStore() *redisStore {
	return &redisStore{
		instances:      map[string]*RedisInstanceEntry{},
		clusters:       map[string]*RedisClusterEntry{},
		instanceErrors: map[string]error{},
		clusterErrors:  map[string]error{},
	}
}

// === Test-side seeding =====================================================

func (s *redisStore) AddRedisInstance(instanceId, instanceClass, engineVersion, status string) *RedisInstanceEntry {
	s.m.Lock()
	defer s.m.Unlock()
	if instanceId == "" {
		instanceId = "r-" + uuid.NewString()[:8]
	}
	entry := &RedisInstanceEntry{
		InstanceId:       instanceId,
		InstanceClass:    instanceClass,
		EngineVersion:    engineVersion,
		InstanceStatus:   status,
		NetworkType:      redisinstance.NetworkType,
		ChargeType:       redisinstance.ChargeType,
		Port:             redisinstance.DefaultPort,
		ConnectionDomain: fmt.Sprintf("%s.redis.rds.aliyuncs.com", instanceId),
	}
	s.instances[instanceId] = entry
	return entry
}

func (s *redisStore) AddRedisCluster(instanceId, instanceClass, engineVersion, status string, shardCount int32) *RedisClusterEntry {
	s.m.Lock()
	defer s.m.Unlock()
	if instanceId == "" {
		instanceId = "r-" + uuid.NewString()[:8]
	}
	entry := &RedisClusterEntry{
		InstanceId:       instanceId,
		InstanceClass:    instanceClass,
		EngineVersion:    engineVersion,
		InstanceStatus:   status,
		ShardCount:       shardCount,
		NetworkType:      redisinstance.NetworkType,
		ChargeType:       redisinstance.ChargeType,
		Port:             redisinstance.DefaultPort,
		ConnectionDomain: fmt.Sprintf("%s.redis.rds.aliyuncs.com", instanceId),
	}
	s.clusters[instanceId] = entry
	return entry
}

func (s *redisStore) SetRedisInstanceError(instanceId string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err == nil {
		delete(s.instanceErrors, instanceId)
	} else {
		s.instanceErrors[instanceId] = err
	}
}

func (s *redisStore) SetRedisClusterError(instanceId string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err == nil {
		delete(s.clusterErrors, instanceId)
	} else {
		s.clusterErrors[instanceId] = err
	}
}

func (s *redisStore) GetRedisInstance(instanceId string) *RedisInstanceEntry {
	s.m.Lock()
	defer s.m.Unlock()
	return s.instances[instanceId]
}

func (s *redisStore) GetRedisCluster(instanceId string) *RedisClusterEntry {
	s.m.Lock()
	defer s.m.Unlock()
	return s.clusters[instanceId]
}

// === Instance-client ops ====================================================

func (s *redisStore) createInstance(ctx context.Context, opts redisinstance.CreateInstanceOptions) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	// If ShardCount>0 the caller wants a cluster; route to cluster store.
	if opts.ShardCount > 0 {
		id := "r-" + uuid.NewString()[:8]
		s.clusters[id] = &RedisClusterEntry{
			InstanceId:       id,
			InstanceName:     opts.InstanceName,
			InstanceClass:    opts.InstanceClass,
			InstanceStatus:   redisinstance.InstanceStatusCreating,
			EngineVersion:    opts.EngineVersion,
			VpcId:            opts.VpcId,
			VSwitchId:        opts.VSwitchId,
			NetworkType:      redisinstance.NetworkType,
			ChargeType:       redisinstance.ChargeType,
			Port:             pickPort(opts.Port),
			ConnectionDomain: fmt.Sprintf("%s.redis.rds.aliyuncs.com", id),
			ShardCount:       opts.ShardCount,
			ReplicasPerShard: opts.ReadOnlyCount,
			Password:         opts.Password,
		}
		return id, nil
	}

	id := "r-" + uuid.NewString()[:8]
	s.instances[id] = &RedisInstanceEntry{
		InstanceId:       id,
		InstanceName:     opts.InstanceName,
		InstanceClass:    opts.InstanceClass,
		InstanceStatus:   redisinstance.InstanceStatusCreating,
		EngineVersion:    opts.EngineVersion,
		VpcId:            opts.VpcId,
		VSwitchId:        opts.VSwitchId,
		NetworkType:      redisinstance.NetworkType,
		ChargeType:       redisinstance.ChargeType,
		Port:             pickPort(opts.Port),
		ConnectionDomain: fmt.Sprintf("%s.redis.rds.aliyuncs.com", id),
		ReadOnlyCount:    opts.ReadOnlyCount,
		Password:         opts.Password,
	}
	return id, nil
}

func pickPort(p int64) int64 {
	if p > 0 {
		return p
	}
	return redisinstance.DefaultPort
}

func (s *redisStore) describeInstanceByName(ctx context.Context, name string) (*redisinstance.InstanceInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	for _, e := range s.instances {
		if e.InstanceName == name {
			return &redisinstance.InstanceInfo{
				InstanceId:       e.InstanceId,
				InstanceStatus:   e.InstanceStatus,
				InstanceClass:    e.InstanceClass,
				ArchitectureType: "standard",
				NetworkType:      e.NetworkType,
				VpcId:            e.VpcId,
				VSwitchId:        e.VSwitchId,
				EngineVersion:    e.EngineVersion,
				Port:             e.Port,
				ConnectionDomain: e.ConnectionDomain,
				ChargeType:       e.ChargeType,
				ReadOnlyCount:    e.ReadOnlyCount,
				Config:           e.Config,
			}, nil
		}
	}
	for _, e := range s.clusters {
		if e.InstanceName == name {
			return &redisinstance.InstanceInfo{
				InstanceId:       e.InstanceId,
				InstanceStatus:   e.InstanceStatus,
				InstanceClass:    e.InstanceClass,
				ArchitectureType: "cluster",
				NetworkType:      e.NetworkType,
				VpcId:            e.VpcId,
				VSwitchId:        e.VSwitchId,
				EngineVersion:    e.EngineVersion,
				Port:             e.Port,
				ConnectionDomain: e.ConnectionDomain,
				ChargeType:       e.ChargeType,
				ShardCount:       e.ShardCount,
				ReadOnlyCount:    e.ReplicasPerShard,
				Config:           e.Config,
			}, nil
		}
	}
	return nil, nil
}

func (s *redisStore) describeInstance(ctx context.Context, instanceId string) (*redisinstance.InstanceInfo, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.instanceErrors[instanceId]; ok {
		return nil, err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return nil, err
	}

	if e := s.instances[instanceId]; e != nil {
		return &redisinstance.InstanceInfo{
			InstanceId:       e.InstanceId,
			InstanceStatus:   e.InstanceStatus,
			InstanceClass:    e.InstanceClass,
			ArchitectureType: "standard",
			NetworkType:      e.NetworkType,
			VpcId:            e.VpcId,
			VSwitchId:        e.VSwitchId,
			EngineVersion:    e.EngineVersion,
			Port:             e.Port,
			ConnectionDomain: e.ConnectionDomain,
			ChargeType:       e.ChargeType,
			ReadOnlyCount:    e.ReadOnlyCount,
			Config:           e.Config,
		}, nil
	}
	if e := s.clusters[instanceId]; e != nil {
		return &redisinstance.InstanceInfo{
			InstanceId:       e.InstanceId,
			InstanceStatus:   e.InstanceStatus,
			InstanceClass:    e.InstanceClass,
			ArchitectureType: "cluster",
			NetworkType:      e.NetworkType,
			VpcId:            e.VpcId,
			VSwitchId:        e.VSwitchId,
			EngineVersion:    e.EngineVersion,
			Port:             e.Port,
			ConnectionDomain: e.ConnectionDomain,
			ChargeType:       e.ChargeType,
			ShardCount:       e.ShardCount,
			ReadOnlyCount:    e.ReplicasPerShard,
			Config:           e.Config,
		}, nil
	}
	return nil, nil
}

func (s *redisStore) modifyInstanceSpec(ctx context.Context, instanceId string, opts redisinstance.ModifyInstanceSpecOptions) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	if e := s.instances[instanceId]; e != nil {
		if opts.InstanceClass != "" {
			e.InstanceClass = opts.InstanceClass
		}
		if opts.ReadOnlyCount != nil {
			e.ReadOnlyCount = *opts.ReadOnlyCount
		}
		e.InstanceStatus = redisinstance.InstanceStatusChanging
		return nil
	}
	if e := s.clusters[instanceId]; e != nil {
		if opts.InstanceClass != "" {
			e.InstanceClass = opts.InstanceClass
		}
		if opts.ShardCount > 0 {
			e.ShardCount = opts.ShardCount
		}
		if opts.ReadOnlyCount != nil {
			e.ReplicasPerShard = *opts.ReadOnlyCount
		}
		e.InstanceStatus = redisinstance.InstanceStatusChanging
		return nil
	}
	return fmt.Errorf("instance %s not found", instanceId)
}

func (s *redisStore) deleteInstance(ctx context.Context, instanceId string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	if _, ok := s.instances[instanceId]; ok {
		delete(s.instances, instanceId)
		return nil
	}
	if _, ok := s.clusters[instanceId]; ok {
		delete(s.clusters, instanceId)
		return nil
	}
	return fmt.Errorf("instance %s not found", instanceId)
}

func (s *redisStore) resetAccountPassword(ctx context.Context, instanceId, accountName, password string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	if e := s.instances[instanceId]; e != nil {
		e.Password = password
		return nil
	}
	if e := s.clusters[instanceId]; e != nil {
		e.Password = password
		return nil
	}
	return fmt.Errorf("instance %s not found", instanceId)
}

func (s *redisStore) modifyInstanceConfig(ctx context.Context, instanceId, config string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	if e := s.instances[instanceId]; e != nil {
		e.Config = config
		return nil
	}
	if e := s.clusters[instanceId]; e != nil {
		e.Config = config
		return nil
	}
	return fmt.Errorf("instance %s not found", instanceId)
}

func (s *redisStore) modifySecurityIps(_ context.Context, instanceId, securityIps string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	if e := s.instances[instanceId]; e != nil {
		e.SecurityIps = securityIps
		return nil
	}
	if e := s.clusters[instanceId]; e != nil {
		e.SecurityIps = securityIps
		return nil
	}
	return fmt.Errorf("instance %s not found", instanceId)
}

func (s *redisStore) describeSecurityIps(_ context.Context, instanceId string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return "", err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return "", err
	}
	if e := s.instances[instanceId]; e != nil {
		return e.SecurityIps, nil
	}
	if e := s.clusters[instanceId]; e != nil {
		return e.SecurityIps, nil
	}
	return "", fmt.Errorf("instance %s not found", instanceId)
}

// === Cluster-client ops (add/delete shards, absolute target) ================

func (s *redisStore) addShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	e := s.clusters[instanceId]
	if e == nil {
		return fmt.Errorf("cluster %s not found", instanceId)
	}
	if targetShardCount <= e.ShardCount {
		return fmt.Errorf("target shard count %d must be greater than current %d", targetShardCount, e.ShardCount)
	}
	e.ShardCount = targetShardCount
	e.InstanceStatus = redisinstance.InstanceStatusChanging
	return nil
}

func (s *redisStore) deleteShardingNode(ctx context.Context, instanceId string, targetShardCount int32) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	e := s.clusters[instanceId]
	if e == nil {
		return fmt.Errorf("cluster %s not found", instanceId)
	}
	if targetShardCount >= e.ShardCount {
		return fmt.Errorf("target shard count %d must be less than current %d", targetShardCount, e.ShardCount)
	}
	if targetShardCount < 1 {
		return fmt.Errorf("target shard count must be >= 1")
	}
	e.ShardCount = targetShardCount
	e.InstanceStatus = redisinstance.InstanceStatusChanging
	return nil
}

// TransitionAllToNormal advances every Creating/Changing/SSLModifying entry to
// Normal. For SSLModifying entries the pending SSL value is also applied.
// Test helper used to simulate the AliCloud API completing an operation.
func (s *redisStore) TransitionAllToNormal() {
	s.m.Lock()
	defer s.m.Unlock()
	for _, e := range s.instances {
		switch e.InstanceStatus {
		case redisinstance.InstanceStatusCreating, redisinstance.InstanceStatusChanging:
			e.InstanceStatus = redisinstance.InstanceStatusNormal
		case redisinstance.InstanceStatusSSLModifying:
			e.InstanceStatus = redisinstance.InstanceStatusNormal
			if e.PendingSslEnabled != nil {
				e.SslEnabled = *e.PendingSslEnabled
				e.PendingSslEnabled = nil
			}
		}
	}
	for _, e := range s.clusters {
		switch e.InstanceStatus {
		case redisinstance.InstanceStatusCreating, redisinstance.InstanceStatusChanging:
			e.InstanceStatus = redisinstance.InstanceStatusNormal
		case redisinstance.InstanceStatusSSLModifying:
			e.InstanceStatus = redisinstance.InstanceStatusNormal
			if e.PendingSslEnabled != nil {
				e.SslEnabled = *e.PendingSslEnabled
				e.PendingSslEnabled = nil
			}
		}
	}
}

func (s *redisStore) describeInstanceSSL(_ context.Context, instanceId string) (bool, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return false, err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return false, err
	}
	if e := s.instances[instanceId]; e != nil {
		// While SSLModifying, return the pending value so callers see the
		// in-flight target and don't retry the operation in a loop.
		if e.PendingSslEnabled != nil {
			return *e.PendingSslEnabled, nil
		}
		return e.SslEnabled, nil
	}
	if e := s.clusters[instanceId]; e != nil {
		if e.PendingSslEnabled != nil {
			return *e.PendingSslEnabled, nil
		}
		return e.SslEnabled, nil
	}
	return false, fmt.Errorf("instance %s not found", instanceId)
}

func (s *redisStore) modifyInstanceSSL(_ context.Context, instanceId string, enable bool) error {
	s.m.Lock()
	defer s.m.Unlock()
	if err, ok := s.instanceErrors[instanceId]; ok {
		return err
	}
	if err, ok := s.clusterErrors[instanceId]; ok {
		return err
	}
	if e := s.instances[instanceId]; e != nil {
		e.PendingSslEnabled = &enable
		e.InstanceStatus = redisinstance.InstanceStatusSSLModifying
		return nil
	}
	if e := s.clusters[instanceId]; e != nil {
		e.PendingSslEnabled = &enable
		e.InstanceStatus = redisinstance.InstanceStatusSSLModifying
		return nil
	}
	return fmt.Errorf("instance %s not found", instanceId)
}

// The client interfaces are satisfied by the redisInstanceClientView and
// redisClusterClientView adapters defined in clientViews.go; see there for the
// var _ = compile-time assertions.
