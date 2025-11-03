package e2e

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/cloud-manager/e2e/sim"
)

type ScenarioSession interface {
	// AddExistingCluster creates new cluster.Cluster on the already created k8s cluster for the given alias.
	// Alias is "kcp", "garden", or the alias an SKR was created with. The created cluster.Cluster is started
	// and stopped when Terminate() is called
	AddExistingCluster(ctx context.Context, alias string) (ClusterInSession, error)

	// CreateNewSkrCluster calls KEB to create new SKR instance and once provisioned it creates new cluster.Cluster
	// that is started. When Terminate() is called the cluster.Cluster is stopped and SKR instance deleted.
	CreateNewSkrCluster(ctx context.Context, opts ...sim.CreateOption) (ClusterInSession, error)

	// AllClusters returns slice of aliases for all clusters, both added and created
	AllClusters() []ClusterInSession

	CurrentCluster() ClusterInSession

	SetCurrentCluster(alias string)

	Terminate(ctx context.Context) error
}

type ClusterInSession interface {
	Cluster
	IsCreatedInSession() bool
	IsCurrent() bool
	RuntimeID() string
	ShootName() string

	AddResources(ctx context.Context, arr ...*ResourceDeclaration) error
}

type defaultClusterInSession struct {
	Cluster
	isCreatedInSession bool
	isCurrent          bool
	runtimeID          string
	shootName          string
	session            ScenarioSession
}

func (c *defaultClusterInSession) AddResources(ctx context.Context, arr ...*ResourceDeclaration) error {
	for _, clstr := range c.session.AllClusters() {
		for _, rd := range arr {
			if clstr.GetResource(rd.Alias) != nil {
				return fmt.Errorf("resource %q already defined", rd.Alias)
			}
		}
	}
	return c.Cluster.AddResources(ctx, arr...)
}

func (c *defaultClusterInSession) IsCreatedInSession() bool {
	return c.isCreatedInSession
}

func (c *defaultClusterInSession) IsCurrent() bool {
	return c.isCurrent
}

func (c *defaultClusterInSession) RuntimeID() string {
	return c.runtimeID
}

func (c *defaultClusterInSession) ShootName() string {
	return c.shootName
}

// CTX ==========================================

type scenarioSessionKeyType struct{}

var scenarioSessionKey = &scenarioSessionKeyType{}

func NewScenarioSession(world WorldIntf) ScenarioSession {
	return &scenarioSession{
		world: world,
	}
}

func StartNewScenarioSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, scenarioSessionKey, NewScenarioSession(GetWorld()))
}

func GetCurrentScenarioSession(ctx context.Context) ScenarioSession {
	val := ctx.Value(scenarioSessionKey)
	if val == nil {
		return nil
	}
	return val.(ScenarioSession)
}

// IMPL ========================================

var _ ScenarioSession = &scenarioSession{}

type scenarioSession struct {
	m sync.Mutex

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	runErr error

	world    WorldIntf
	clusters []*defaultClusterInSession

	terminated bool
}

func (s *scenarioSession) AddExistingCluster(ctx context.Context, alias string) (ClusterInSession, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.terminated {
		return nil, errors.New("can not add cluster to terminated session")
	}

	for _, c := range s.clusters {
		if c.ClusterAlias() == alias {
			return nil, fmt.Errorf("cluster %s already added to scenario", alias)
		}
	}

	if alias == s.world.Kcp().ClusterAlias() {
		cc := &defaultClusterInSession{
			Cluster:            s.world.Kcp(),
			isCreatedInSession: false,
			session:            s,
		}
		s.clusters = append(s.clusters, cc)
		s.SetCurrentCluster(alias)
		return cc, nil
	}

	if alias == s.world.Garden().ClusterAlias() {
		cc := &defaultClusterInSession{
			Cluster:            s.world.Garden(),
			isCreatedInSession: false,
			session:            s,
		}
		s.clusters = append(s.clusters, cc)
		s.SetCurrentCluster(alias)
		return cc, nil
	}

	arr, err := s.world.Sim().Keb().List(ctx, sim.WithAlias(alias))
	if err != nil {
		return nil, fmt.Errorf("error listing runtimes from KEB: %w", err)
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("runtimes with alias %q not found", alias)
	}
	if len(arr) > 1 {
		return nil, fmt.Errorf("found more then one runtime with alias %q", alias)
	}

	id := &arr[0]

	if !id.ProvisioningCompleted {
		err = s.world.Sim().Keb().WaitProvisioningCompleted(ctx, sim.WithRuntime(id.RuntimeID), sim.WithTimeout(15*time.Minute))
		if err != nil {
			return nil, err
		}
	}

	return s.createManagerAndStartIt(ctx, id, false)
}

func (s *scenarioSession) CreateNewSkrCluster(ctx context.Context, opts ...sim.CreateOption) (ClusterInSession, error) {
	var alias string
	for _, o := range opts {
		if aa, ok := o.(sim.WithAlias); ok {
			alias = string(aa)
		}
	}
	if alias == "" {
		return nil, errors.New("must specify an alias of the new skr cluster")
	}

	s.m.Lock()
	defer s.m.Unlock()

	if s.terminated {
		return nil, errors.New("can not add cluster to terminated session")
	}

	for _, c := range s.clusters {
		if c.ClusterAlias() == alias {
			return nil, fmt.Errorf("cluster %q already added to scenario", alias)
		}
	}

	id, err := s.world.Sim().Keb().CreateInstance(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating runtime in KEB: %w", err)
	}

	err = s.world.Sim().Keb().WaitProvisioningCompleted(ctx, sim.WithRuntime(id.RuntimeID), sim.WithTimeout(15*time.Minute))
	if err != nil {
		return nil, err
	}

	return s.createManagerAndStartIt(ctx, &id, true)
}

func (s *scenarioSession) createManagerAndStartIt(ctx context.Context, id *sim.InstanceDetails, isCreatedInSession bool) (ClusterInSession, error) {
	clstr, err := s.world.Sim().CreateSkrManager(ctx, id.RuntimeID)
	if err != nil {
		return nil, fmt.Errorf("error creating client cluster for runtime: %w", err)
	}

	cc := &defaultClusterInSession{
		Cluster:            NewCluster(ctx, id.Alias, clstr),
		isCreatedInSession: isCreatedInSession,
		runtimeID:          id.RuntimeID,
		shootName:          id.ShootName,
		session:            s,
	}
	s.clusters = append(s.clusters, cc)
	s.SetCurrentCluster(id.Alias)

	s.wg.Add(1)
	if s.ctx == nil {
		s.ctx, s.cancel = context.WithCancel(ctx)
	}
	go func() {
		defer s.wg.Done()
		if err := cc.Start(s.ctx); err != nil {
			s.runErr = multierror.Append(s.runErr, fmt.Errorf("error running %q: %w", id.Alias, err))
		}
	}()

	return cc, nil
}

func (s *scenarioSession) AllClusters() []ClusterInSession {
	result := make([]ClusterInSession, len(s.clusters))
	for i, v := range s.clusters {
		result[i] = ClusterInSession(v)
	}
	return result
}

func (s *scenarioSession) CurrentCluster() ClusterInSession {
	for _, c := range s.clusters {
		if c.IsCurrent() {
			return c
		}
	}
	return nil
}

func (s *scenarioSession) SetCurrentCluster(alias string) {
	for _, c := range s.clusters {
		c.isCurrent = c.ClusterAlias() == alias
	}
}

func (s *scenarioSession) Terminate(ctx context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()

	if s.terminated {
		return fmt.Errorf("already terminated")
	}
	s.terminated = true

	s.cancel()
	s.wg.Wait()

	for _, c := range s.clusters {
		if c.IsCreatedInSession() {
			if err := s.world.Sim().Keb().DeleteInstance(ctx, sim.WithRuntime(c.RuntimeID())); err != nil {
				s.runErr = multierror.Append(s.runErr, fmt.Errorf("error deleting cluster %q: %w", c.ClusterAlias(), err))
			}
		}
	}

	return s.runErr
}
