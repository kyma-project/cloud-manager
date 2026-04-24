package mock

import (
	"context"
	"errors"
	"strings"
	"sync"

	logstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"k8s.io/utils/ptr"
)

type logGroupEntry struct {
	name          string
	retentionDays *int32
	tags          map[string]string
}

type logsStore struct {
	m      sync.Mutex
	groups map[string]*logGroupEntry
}

func newLogsStore() *logsStore {
	return &logsStore{
		groups: make(map[string]*logGroupEntry),
	}
}

func (s *logsStore) CreateLogGroup(ctx context.Context, logGroupName string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if _, exists := s.groups[logGroupName]; exists {
		return nil // Idempotent - already exists
	}

	s.groups[logGroupName] = &logGroupEntry{
		name: logGroupName,
		tags: make(map[string]string),
	}
	return nil
}

func (s *logsStore) DeleteLogGroup(ctx context.Context, logGroupName string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if _, exists := s.groups[logGroupName]; !exists {
		return nil // Idempotent - already deleted
	}

	delete(s.groups, logGroupName)
	return nil
}

func (s *logsStore) GetLogGroup(ctx context.Context, logGroupName string) (*logstypes.LogGroup, error) {
	s.m.Lock()
	defer s.m.Unlock()

	entry, exists := s.groups[logGroupName]
	if !exists {
		return nil, &logstypes.ResourceNotFoundException{
			Message: ptr.To("Log group not found"),
		}
	}

	return &logstypes.LogGroup{
		LogGroupName:    ptr.To(entry.name),
		RetentionInDays: entry.retentionDays,
	}, nil
}

func (s *logsStore) DescribeLogGroups(ctx context.Context, prefix string) ([]logstypes.LogGroup, error) {
	s.m.Lock()
	defer s.m.Unlock()

	result := make([]logstypes.LogGroup, 0)
	for name, entry := range s.groups {
		if strings.HasPrefix(name, prefix) {
			lg := logstypes.LogGroup{
				LogGroupName:    ptr.To(name),
				RetentionInDays: entry.retentionDays,
			}
			result = append(result, lg)
		}
	}
	return result, nil
}

func (s *logsStore) PutRetentionPolicy(ctx context.Context, logGroupName string, retentionDays int32) error {
	s.m.Lock()
	defer s.m.Unlock()

	entry, exists := s.groups[logGroupName]
	if !exists {
		return errors.New("log group not found")
	}

	entry.retentionDays = ptr.To(retentionDays)
	return nil
}

func (s *logsStore) TagLogGroup(ctx context.Context, logGroupName string, tags map[string]string) error {
	s.m.Lock()
	defer s.m.Unlock()

	entry, exists := s.groups[logGroupName]
	if !exists {
		return errors.New("log group not found")
	}

	for k, v := range tags {
		entry.tags[k] = v
	}
	return nil
}
