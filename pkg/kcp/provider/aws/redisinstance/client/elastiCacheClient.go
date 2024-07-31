package client

import (
	"context"
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticacheTypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"

	"k8s.io/utils/ptr"
)

func NewClientProvider() awsclient.SkrClientProvider[ElastiCacheClient] {
	return func(ctx context.Context, region, key, secret, role string) (ElastiCacheClient, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		return newClient(
			ec2.NewFromConfig(cfg),
			elasticache.NewFromConfig(cfg),
		), nil
	}
}

type CreateElastiCacheClusterOptions struct {
	Name                    string
	SubnetGroupName         string
	ParameterGroupName      string
	CacheNodeType           string
	EngineVersion           string
	AutoMinorVersionUpgrade bool
}

type ElastiCacheClient interface {
	DescribeElastiCacheSubnetGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheSubnetGroup, error)
	CreateElastiCacheSubnetGroup(ctx context.Context, name string, subnetIds []string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheSubnetGroupOutput, error)
	DeleteElastiCacheSubnetGroup(ctx context.Context, name string) error

	DescribeElastiCacheParameterGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheParameterGroup, error)
	CreateElastiCacheParameterGroup(ctx context.Context, name, family string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheParameterGroupOutput, error)
	DeleteElastiCacheParameterGroup(ctx context.Context, name string) error
	DescribeElastiCacheParameters(ctx context.Context, groupName string) ([]elasticacheTypes.Parameter, error)
	ModifyElastiCacheParameterGroup(ctx context.Context, groupName string, parameters []elasticacheTypes.ParameterNameValue) error
	DescribeEngineDefaultParameters(ctx context.Context, family string) ([]elasticacheTypes.Parameter, error)

	DescribeElastiCacheCluster(ctx context.Context, clusterId string) ([]elasticacheTypes.CacheCluster, error)
	CreateElastiCacheCluster(ctx context.Context, tags []elasticacheTypes.Tag, options CreateElastiCacheClusterOptions) (*elasticache.CreateCacheClusterOutput, error)
	DeleteElastiCacheClaster(ctx context.Context, id string) error
}

func newClient(ec2Svc *ec2.Client, elastiCacheSvc *elasticache.Client) ElastiCacheClient {
	return &client{
		ec2Svc:         ec2Svc,
		elastiCacheSvc: elastiCacheSvc,
	}
}

type client struct {
	ec2Svc         *ec2.Client
	elastiCacheSvc *elasticache.Client
}

func (c *client) DescribeElastiCacheSubnetGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheSubnetGroup, error) {

	out, err := c.elastiCacheSvc.DescribeCacheSubnetGroups(ctx, &elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: ptr.To(name),
	})
	if err != nil {
		if awsmeta.IsNotFound(err) {
			return []elasticacheTypes.CacheSubnetGroup{}, nil
		}

		return nil, err
	}

	return out.CacheSubnetGroups, nil
}

func (c *client) CreateElastiCacheSubnetGroup(ctx context.Context, name string, subnetIds []string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheSubnetGroupOutput, error) {
	out, err := c.elastiCacheSvc.CreateCacheSubnetGroup(ctx, &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: ptr.To(fmt.Sprintf("SubnetGroup for ElastiCache %s", name)),
		CacheSubnetGroupName:        ptr.To(name),
		Tags:                        tags,
		SubnetIds:                   subnetIds,
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *client) DeleteElastiCacheSubnetGroup(ctx context.Context, name string) error {
	_, err := c.elastiCacheSvc.DeleteCacheSubnetGroup(ctx, &elasticache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: ptr.To(name),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *client) DescribeElastiCacheParameterGroup(ctx context.Context, name string) ([]elasticacheTypes.CacheParameterGroup, error) {
	out, err := c.elastiCacheSvc.DescribeCacheParameterGroups(ctx, &elasticache.DescribeCacheParameterGroupsInput{
		CacheParameterGroupName: ptr.To(name),
	})
	if err != nil {
		if awsmeta.IsNotFound(err) {
			return []elasticacheTypes.CacheParameterGroup{}, nil
		}

		return nil, err
	}

	return out.CacheParameterGroups, nil
}

func (c *client) CreateElastiCacheParameterGroup(ctx context.Context, name, family string, tags []elasticacheTypes.Tag) (*elasticache.CreateCacheParameterGroupOutput, error) {
	out, err := c.elastiCacheSvc.CreateCacheParameterGroup(ctx, &elasticache.CreateCacheParameterGroupInput{
		CacheParameterGroupName:   ptr.To(name),
		CacheParameterGroupFamily: ptr.To(family),
		Tags:                      tags,
		Description:               ptr.To(fmt.Sprintf("ParameterGroup for ElastiCache %s", name)),
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *client) DeleteElastiCacheParameterGroup(ctx context.Context, name string) error {
	_, err := c.elastiCacheSvc.DeleteCacheParameterGroup(ctx, &elasticache.DeleteCacheParameterGroupInput{
		CacheParameterGroupName: ptr.To(name),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *client) DescribeElastiCacheParameters(ctx context.Context, groupName string) ([]elasticacheTypes.Parameter, error) {
	result := []elasticacheTypes.Parameter{}
	var marker *string = nil
	for {
		out, err := c.elastiCacheSvc.DescribeCacheParameters(ctx, &elasticache.DescribeCacheParametersInput{
			CacheParameterGroupName: ptr.To(groupName),
			Marker:                  marker,
		})

		if err != nil {
			return nil, err
		}

		result = append(result, out.Parameters...)
		if out.Marker == nil {
			break
		}
	}

	return result, nil
}

func (c *client) ModifyElastiCacheParameterGroup(ctx context.Context, groupName string, parameters []elasticacheTypes.ParameterNameValue) error {
	apiMaxChunkSize := 20
	for i := 0; i < len(parameters); i += apiMaxChunkSize {
		end := int(math.Min(float64(i+apiMaxChunkSize), float64(len(parameters))))

		chunk := parameters[i:end]

		_, err := c.elastiCacheSvc.ModifyCacheParameterGroup(ctx, &elasticache.ModifyCacheParameterGroupInput{
			CacheParameterGroupName: ptr.To(groupName),
			ParameterNameValues:     chunk,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

var defaultParametersCache map[string][]elasticacheTypes.Parameter = map[string][]elasticacheTypes.Parameter{}

func (c *client) DescribeEngineDefaultParameters(ctx context.Context, family string) ([]elasticacheTypes.Parameter, error) {
	if _, found := defaultParametersCache[family]; found {
		return defaultParametersCache[family], nil
	}

	result := []elasticacheTypes.Parameter{}
	var marker *string = nil
	for {
		out, err := c.elastiCacheSvc.DescribeEngineDefaultParameters(ctx, &elasticache.DescribeEngineDefaultParametersInput{
			CacheParameterGroupFamily: ptr.To(family),
			Marker:                    marker,
		})

		if err != nil {
			return nil, err
		}

		result = append(result, out.EngineDefaults.Parameters...)
		if out.EngineDefaults.Marker == nil {
			break
		}
	}

	defaultParametersCache[family] = result

	return result, nil
}

func (c *client) DescribeElastiCacheCluster(ctx context.Context, clusterId string) ([]elasticacheTypes.CacheCluster, error) {
	out, err := c.elastiCacheSvc.DescribeCacheClusters(ctx, &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    ptr.To(clusterId),
		ShowCacheNodeInfo: ptr.To(true),
	})
	if err != nil {
		if awsmeta.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return out.CacheClusters, nil
}

func (c *client) CreateElastiCacheCluster(ctx context.Context, tags []elasticacheTypes.Tag, options CreateElastiCacheClusterOptions) (*elasticache.CreateCacheClusterOutput, error) {
	params := &elasticache.CreateCacheClusterInput{
		CacheClusterId:          aws.String(options.Name),
		CacheSubnetGroupName:    aws.String(options.SubnetGroupName),
		CacheParameterGroupName: aws.String(options.ParameterGroupName),
		CacheNodeType:           aws.String(options.CacheNodeType),
		NumCacheNodes:           aws.Int32(1),
		Engine:                  aws.String("redis"),
		EngineVersion:           aws.String(options.EngineVersion),
		AutoMinorVersionUpgrade: aws.Bool(options.AutoMinorVersionUpgrade),
		Tags:                    tags,
	}

	res, err := c.elastiCacheSvc.CreateCacheCluster(ctx, params)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *client) DeleteElastiCacheClaster(ctx context.Context, id string) error {
	deleteInput := &elasticache.DeleteCacheClusterInput{
		CacheClusterId: ptr.To(id),
	}

	_, err := c.elastiCacheSvc.DeleteCacheCluster(ctx, deleteInput)

	return err
}
