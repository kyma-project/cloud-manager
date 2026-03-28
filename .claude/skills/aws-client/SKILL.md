---
name: aws-client
description: Implement AWS cloud provider clients in Cloud Manager. Use when creating AWS API clients, working with AWS resources, or implementing AWS-specific reconcilers.
---

# AWS Client Implementation

Create AWS cloud provider clients for Cloud Manager.

## Quick Start

1. Create client interface in `pkg/kcp/provider/aws/<resource>/client/`
2. Implement using AWS SDK for Go v2
3. Create provider function
4. Wire in `cmd/main.go`

## Client Interface Pattern

**File**: `pkg/kcp/provider/aws/<resource>/client/client.go`

```go
package client

import (
    "context"
    "github.com/aws/aws-sdk-for-go-v2/service/elasticache"
)

// Business operations interface
type ElastiCacheClient interface {
    CreateReplicationGroup(ctx context.Context, req CreateReplicationGroupRequest) (string, error)
    DescribeReplicationGroup(ctx context.Context, id string) (*ReplicationGroup, error)
    ModifyReplicationGroup(ctx context.Context, id string, req ModifyRequest) error
    DeleteReplicationGroup(ctx context.Context, id string) error
}

type CreateReplicationGroupRequest struct {
    ReplicationGroupId          string
    ReplicationGroupDescription string
    Engine                      string
    EngineVersion               string
    NodeType                    string
    NumNodeGroups               int32
    ReplicasPerNodeGroup        int32
}

type ReplicationGroup struct {
    ReplicationGroupId string
    Status             string
    PrimaryEndpoint    string
    Port               int32
}
```

## Implementation

```go
type elastiCacheClient struct {
    client *elasticache.Client
}

func NewElastiCacheClient(cfg aws.Config) ElastiCacheClient {
    return &elastiCacheClient{
        client: elasticache.NewFromConfig(cfg),
    }
}

func (c *elastiCacheClient) CreateReplicationGroup(ctx context.Context, req CreateReplicationGroupRequest) (string, error) {
    input := &elasticache.CreateReplicationGroupInput{
        ReplicationGroupId:          aws.String(req.ReplicationGroupId),
        ReplicationGroupDescription: aws.String(req.ReplicationGroupDescription),
        Engine:                      aws.String(req.Engine),
        EngineVersion:               aws.String(req.EngineVersion),
        CacheNodeType:               aws.String(req.NodeType),
        NumNodeGroups:               aws.Int32(req.NumNodeGroups),
        ReplicasPerNodeGroup:        aws.Int32(req.ReplicasPerNodeGroup),
    }

    _, err := c.client.CreateReplicationGroup(ctx, input)
    if err != nil {
        return "", err
    }

    return req.ReplicationGroupId, nil
}

func (c *elastiCacheClient) DescribeReplicationGroup(ctx context.Context, id string) (*ReplicationGroup, error) {
    input := &elasticache.DescribeReplicationGroupsInput{
        ReplicationGroupId: aws.String(id),
    }

    output, err := c.client.DescribeReplicationGroups(ctx, input)
    if err != nil {
        return nil, err
    }

    if len(output.ReplicationGroups) == 0 {
        return nil, &NotFoundError{Id: id}
    }

    rg := output.ReplicationGroups[0]
    return &ReplicationGroup{
        ReplicationGroupId: aws.ToString(rg.ReplicationGroupId),
        Status:             aws.ToString(rg.Status),
        PrimaryEndpoint:    aws.ToString(rg.NodeGroups[0].PrimaryEndpoint.Address),
        Port:               aws.ToInt32(rg.NodeGroups[0].PrimaryEndpoint.Port),
    }, nil
}

func (c *elastiCacheClient) DeleteReplicationGroup(ctx context.Context, id string) error {
    input := &elasticache.DeleteReplicationGroupInput{
        ReplicationGroupId: aws.String(id),
    }

    _, err := c.client.DeleteReplicationGroup(ctx, input)
    return err
}
```

## Error Handling

```go
import (
    "github.com/aws/aws-sdk-for-go-v2/service/elasticache/types"
    "github.com/aws/smithy-go"
)

type NotFoundError struct {
    Id string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("resource %s not found", e.Id)
}

func IsNotFound(err error) bool {
    var notFound *NotFoundError
    if errors.As(err, &notFound) {
        return true
    }

    var apiErr smithy.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.ErrorCode() {
        case "ReplicationGroupNotFoundFault",
             "CacheClusterNotFound",
             "InvalidParameterValue":
            return true
        }
    }
    return false
}

// Usage
rg, err := client.DescribeReplicationGroup(ctx, id)
if err != nil {
    if IsNotFound(err) {
        return nil, ctx
    }
    return err, ctx
}
```

## Provider Pattern

```go
// Provider function
func NewElastiCacheClientProvider(cfg aws.Config) func() ElastiCacheClient {
    return func() ElastiCacheClient {
        return NewElastiCacheClient(cfg)
    }
}

// In state factory
type stateFactory struct {
    elastiCacheClientProvider func() ElastiCacheClient
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
    return &State{
        State:            focalState,
        elastiCacheClient: f.elastiCacheClientProvider(),
    }, nil
}
```

## Async Operations

AWS operations are eventually consistent. Check status:

```go
func waitReplicationGroupAvailable(ctx context.Context, st composed.State) (error, context.Context) {
    state := st.(*State)
    obj := state.ObjAsAwsRedisCluster()

    if obj.Status.Id == "" {
        return nil, ctx
    }

    rg, err := state.elastiCacheClient.DescribeReplicationGroup(ctx, obj.Status.Id)
    if err != nil {
        if IsNotFound(err) {
            return nil, ctx
        }
        return err, ctx
    }

    switch rg.Status {
    case "available":
        state.replicationGroup = rg
        return nil, ctx
    case "create-failed", "deleting":
        return fmt.Errorf("replication group status: %s", rg.Status), ctx
    default:
        // creating, modifying, etc.
        return composed.StopWithRequeueDelay(30 * time.Second), nil
    }
}
```

## AWS Config Setup

```go
// In main.go
cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRegion(region),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        accessKeyId, secretAccessKey, "",
    )),
)
if err != nil {
    setupLog.Error(err, "failed to load AWS config")
    os.Exit(1)
}

elastiCacheClientProvider := awsclient.NewElastiCacheClientProvider(cfg)
```

## Checklist

- [ ] Client interface defined
- [ ] Implementation uses AWS SDK v2
- [ ] Error handling with IsNotFound
- [ ] Provider function created
- [ ] Status polling for async operations
- [ ] AWS config setup in main.go

## Related

- Add reconciler: `/add-kcp-reconciler`
- AWS SDK v2: https://aws.github.io/aws-sdk-go-v2/
