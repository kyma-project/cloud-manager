package mock

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

type CloudWatchLogsConfig interface {
	SetLogGroupError(logGroupArn string, err error)
}

type logGroupEntry struct {
	logGroup cloudwatchlogstypes.LogGroup
	tags     map[string]string
}

type cloudWatchLogsStore struct {
	m        sync.Mutex
	account  string
	region   string
	items    []*logGroupEntry
	errorMap map[string]error
}

var _ awsclient.CloudWatchLogsClient = (*cloudWatchLogsStore)(nil)
var _ CloudWatchLogsConfig = (*cloudWatchLogsStore)(nil)

func newCloudWatchLogsStore(account, region string) *cloudWatchLogsStore {
	return &cloudWatchLogsStore{
		account:  account,
		region:   region,
		errorMap: make(map[string]error),
	}
}

// buildLogGroupArn returns the no-trailing-:* ARN used as ResourceArn in tag APIs.
func (s *cloudWatchLogsStore) buildLogGroupArn(logGroupName string) string {
	a := &arn.ARN{
		Partition: awsconfig.AwsConfig.ArnPartition,
		Service:   "logs",
		Region:    s.region,
		AccountID: s.account,
		Resource:  fmt.Sprintf("log-group:%s", logGroupName),
	}
	return a.String()
}

// arnWithWildcard returns the Arn field value (with trailing :*).
func arnWithWildcard(logGroupArn string) string {
	return logGroupArn + ":*"
}

func (s *cloudWatchLogsStore) findByName(name string) *logGroupEntry {
	for _, e := range s.items {
		if ptr.Deref(e.logGroup.LogGroupName, "") == name {
			return e
		}
	}
	return nil
}

func (s *cloudWatchLogsStore) findByArn(resourceArn string) *logGroupEntry {
	// ResourceArn may be with or without trailing :*
	normalized := strings.TrimSuffix(resourceArn, ":*")
	for _, e := range s.items {
		if ptr.Deref(e.logGroup.LogGroupArn, "") == normalized {
			return e
		}
	}
	return nil
}

// CloudWatchLogsConfig implementation =========================================

func (s *cloudWatchLogsStore) SetLogGroupError(logGroupArn string, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if err == nil {
		delete(s.errorMap, logGroupArn)
	} else {
		s.errorMap[logGroupArn] = err
	}
}

// CloudWatchLogsClient implementation =========================================

func (s *cloudWatchLogsStore) CreateLogGroup(ctx context.Context, params *cloudwatchlogs.CreateLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.CreateLogGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	name := ptr.Deref(params.LogGroupName, "")
	logGroupArn := s.buildLogGroupArn(name)

	if err, ok := s.errorMap[logGroupArn]; ok {
		return nil, err
	}

	if s.findByName(name) != nil {
		return nil, &cloudwatchlogstypes.ResourceAlreadyExistsException{
			Message: new(fmt.Sprintf("log group %q already exists", name)),
		}
	}

	now := time.Now().UnixMilli()
	entry := &logGroupEntry{
		logGroup: cloudwatchlogstypes.LogGroup{
			LogGroupName:  new(name),
			LogGroupArn:   new(logGroupArn),
			Arn:           new(arnWithWildcard(logGroupArn)),
			CreationTime:  new(now),
			LogGroupClass: params.LogGroupClass,
		},
		tags: make(map[string]string),
	}

	if params.KmsKeyId != nil {
		entry.logGroup.KmsKeyId = new(ptr.Deref(params.KmsKeyId, ""))
	}
	if params.DeletionProtectionEnabled != nil {
		entry.logGroup.DeletionProtectionEnabled = new(ptr.Deref(params.DeletionProtectionEnabled, false))
	}

	maps.Copy(entry.tags, params.Tags)

	s.items = append(s.items, entry)

	return &cloudwatchlogs.CreateLogGroupOutput{}, nil
}

func (s *cloudWatchLogsStore) DeleteLogGroup(ctx context.Context, params *cloudwatchlogs.DeleteLogGroupInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DeleteLogGroupOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	name := ptr.Deref(params.LogGroupName, "")
	logGroupArn := s.buildLogGroupArn(name)

	if err, ok := s.errorMap[logGroupArn]; ok {
		return nil, err
	}

	idx := -1
	for i, e := range s.items {
		if ptr.Deref(e.logGroup.LogGroupName, "") == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, &cloudwatchlogstypes.ResourceNotFoundException{
			Message: new(fmt.Sprintf("log group %q not found", name)),
		}
	}

	s.items = append(s.items[:idx], s.items[idx+1:]...)

	return &cloudwatchlogs.DeleteLogGroupOutput{}, nil
}

func (s *cloudWatchLogsStore) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	prefix := ptr.Deref(params.LogGroupNamePrefix, "")
	pattern := ptr.Deref(params.LogGroupNamePattern, "")

	var matched []cloudwatchlogstypes.LogGroup
	for _, e := range s.items {
		name := ptr.Deref(e.logGroup.LogGroupName, "")

		if len(params.LogGroupIdentifiers) > 0 {
			found := false
			lgArn := ptr.Deref(e.logGroup.LogGroupArn, "")
			for _, id := range params.LogGroupIdentifiers {
				normalized := strings.TrimSuffix(id, ":*")
				if normalized == lgArn || id == name {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		} else if prefix != "" {
			// logGroupNamePrefix and logGroupNamePattern are mutually exclusive
			if !strings.HasPrefix(name, prefix) {
				continue
			}
		} else if pattern != "" {
			if !strings.Contains(name, pattern) {
				continue
			}
		}

		cpy, err := util.JsonClone(&e.logGroup)
		if err != nil {
			return nil, err
		}
		matched = append(matched, *cpy)
	}

	return &cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: matched,
	}, nil
}

func (s *cloudWatchLogsStore) PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutRetentionPolicyOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	name := ptr.Deref(params.LogGroupName, "")
	if name == "" {
		return nil, fmt.Errorf("log group name is required")
	}

	retention := ptr.Deref(params.RetentionInDays, 0)

	for _, item := range s.items {
		if ptr.Deref(item.logGroup.LogGroupName, "") == name {
			item.logGroup.RetentionInDays = new(retention)
			return &cloudwatchlogs.PutRetentionPolicyOutput{}, nil
		}
	}

	return nil, &cloudwatchlogstypes.ResourceNotFoundException{
		Message: new(fmt.Sprintf("log group %q not found", name)),
	}
}

func (s *cloudWatchLogsStore) ListTagsForResource(ctx context.Context, params *cloudwatchlogs.ListTagsForResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListTagsForResourceOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	resourceArn := ptr.Deref(params.ResourceArn, "")

	if err, ok := s.errorMap[strings.TrimSuffix(resourceArn, ":*")]; ok {
		return nil, err
	}

	entry := s.findByArn(resourceArn)
	if entry == nil {
		return nil, &cloudwatchlogstypes.ResourceNotFoundException{
			Message: new(fmt.Sprintf("log group with ARN %q not found", resourceArn)),
		}
	}

	tagsCopy, err := util.JsonClone(entry.tags)
	if err != nil {
		return nil, err
	}

	return &cloudwatchlogs.ListTagsForResourceOutput{
		Tags: tagsCopy,
	}, nil
}

func (s *cloudWatchLogsStore) TagResource(ctx context.Context, params *cloudwatchlogs.TagResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.TagResourceOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	resourceArn := ptr.Deref(params.ResourceArn, "")

	if err, ok := s.errorMap[strings.TrimSuffix(resourceArn, ":*")]; ok {
		return nil, err
	}

	entry := s.findByArn(resourceArn)
	if entry == nil {
		return nil, &cloudwatchlogstypes.ResourceNotFoundException{
			Message: new(fmt.Sprintf("log group with ARN %q not found", resourceArn)),
		}
	}

	maps.Copy(entry.tags, params.Tags)

	return &cloudwatchlogs.TagResourceOutput{}, nil
}

func (s *cloudWatchLogsStore) UntagResource(ctx context.Context, params *cloudwatchlogs.UntagResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.UntagResourceOutput, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}
	s.m.Lock()
	defer s.m.Unlock()

	resourceArn := ptr.Deref(params.ResourceArn, "")

	if err, ok := s.errorMap[strings.TrimSuffix(resourceArn, ":*")]; ok {
		return nil, err
	}

	entry := s.findByArn(resourceArn)
	if entry == nil {
		return nil, &cloudwatchlogstypes.ResourceNotFoundException{
			Message: new(fmt.Sprintf("log group with ARN %q not found", resourceArn)),
		}
	}

	for _, k := range params.TagKeys {
		delete(entry.tags, k)
	}

	return &cloudwatchlogs.UntagResourceOutput{}, nil
}
