package e2e

import (
	"context"
	"errors"
)

type ScenarioSession interface {
	RegisterCluster(alias string)
	AllRegisteredClusters() []string

	CurrentCluster() SkrCluster
	CurrentClusterAlias() string
	SetCurrentCluster(c SkrCluster, alias string)
}

// CTX ==========================================

type scenarioSessionKeyType struct{}

var scenarioSessionKey = &scenarioSessionKeyType{}

func NewScenarioSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, scenarioSessionKey, &scenarioSession{})
}

func GetScenarioSession(ctx context.Context) ScenarioSession {
	val := ctx.Value(scenarioSessionKey)
	if val == nil {
		return nil
	}
	return val.(ScenarioSession)
}

func GetScenarioSessionEnsureCluster(ctx context.Context) (ScenarioSession, error) {
	session := GetScenarioSession(ctx)
	if session == nil {
		return nil, errors.New("scenario session not started")
	}
	if session.CurrentCluster() == nil {
		return nil, errors.New("scenario session does not have cluster set")
	}
	return session, nil
}

// IMPL ========================================

var _ ScenarioSession = &scenarioSession{}

type scenarioSession struct {
	transientClusters   []string
	currentCluster      SkrCluster
	currentClusterAlias string
}

func (s *scenarioSession) RegisterCluster(alias string) {
	s.transientClusters = append(s.transientClusters, alias)
}

func (s *scenarioSession) AllRegisteredClusters() []string {
	return append(make([]string, 0, len(s.transientClusters)), s.transientClusters...)
}

func (s *scenarioSession) CurrentCluster() SkrCluster {
	return s.currentCluster
}

func (s *scenarioSession) CurrentClusterAlias() string {
	return s.currentClusterAlias
}

func (s *scenarioSession) SetCurrentCluster(c SkrCluster, alias string) {
	s.currentCluster = c
	s.currentClusterAlias = alias
}
