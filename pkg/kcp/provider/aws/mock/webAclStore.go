package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/aws/smithy-go"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

type WebAclConfig interface {
	SetWebAclError(id string, err error)
	InitiateWebAcl(id, name string, scope types.Scope)
}

type webAclEntry struct {
	webAcl    types.WebACL
	lockToken string
}

type webAclStore struct {
	m        sync.Mutex
	items    []*webAclEntry
	errorMap map[string]error
	account  string
	region   string
}

func newWebAclStore(account, region string) *webAclStore {
	return &webAclStore{
		errorMap: make(map[string]error),
		account:  account,
		region:   region,
	}
}

func (s *webAclStore) CreateWebACL(ctx context.Context, input *wafv2.CreateWebACLInput) (*types.WebACL, string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	name := ptr.Deref(input.Name, "")
	description := ptr.Deref(input.Description, "")

	// Check if WebACL with this name already exists
	for _, x := range s.items {
		if ptr.Equal(x.webAcl.Name, input.Name) {
			return nil, "", &smithy.GenericAPIError{
				Code:    "WAFDuplicateItemException",
				Message: fmt.Sprintf("WebACL with name %s already exists", name),
			}
		}
	}

	id := uuid.NewString()
	arn := awsutil.Waf2Arn(s.region, s.account, name, id)
	lockToken := uuid.NewString()

	// Deep copy inputs to avoid shared references
	defaultActionCopy, err := util.JsonClone(input.DefaultAction)
	if err != nil {
		return nil, "", err
	}
	rulesCopy, err := util.JsonClone(input.Rules)
	if err != nil {
		return nil, "", err
	}
	visibilityConfigCopy, err := util.JsonClone(input.VisibilityConfig)
	if err != nil {
		return nil, "", err
	}
	customResponseBodiesCopy, err := util.JsonClone(input.CustomResponseBodies)
	if err != nil {
		return nil, "", err
	}

	webAcl := types.WebACL{
		Id:                   new(id),
		Name:                 new(name),
		ARN:                  new(arn),
		Description:          new(description),
		DefaultAction:        defaultActionCopy,
		Rules:                rulesCopy,
		VisibilityConfig:     visibilityConfigCopy,
		CustomResponseBodies: customResponseBodiesCopy,
		Capacity:             100,
	}

	item := &webAclEntry{
		webAcl:    webAcl,
		lockToken: lockToken,
	}

	s.items = append(s.items, item)

	webAclCopy, err := util.JsonClone(&webAcl)
	if err != nil {
		return nil, "", err
	}

	return webAclCopy, lockToken, nil
}

func (s *webAclStore) GetWebACL(ctx context.Context, name, id string, scope types.Scope) (*types.WebACL, string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[id]; ok && err != nil {
		return nil, "", err
	}

	entry, err := s.findWebAclEntry(name, id)
	if err != nil {
		return nil, "", err
	}

	webAclCopy, err := util.JsonClone(&entry.webAcl)
	if err != nil {
		return nil, "", err
	}

	return webAclCopy, entry.lockToken, nil
}

func (s *webAclStore) findWebAclEntry(name, id string) (*webAclEntry, error) {
	for _, x := range s.items {
		if ptr.Equal(x.webAcl.Id, new(id)) || ptr.Equal(x.webAcl.Name, new(name)) {
			return x, nil
		}
	}

	return nil, &smithy.GenericAPIError{
		Code:    "WAFNonexistentItemException",
		Message: fmt.Sprintf("WebACL with id %s or name %s does not exist", id, name),
	}
}

func (s *webAclStore) UpdateWebACL(ctx context.Context, input *wafv2.UpdateWebACLInput) error {
	s.m.Lock()
	defer s.m.Unlock()

	id := ptr.Deref(input.Id, "")
	lockToken := ptr.Deref(input.LockToken, "")

	if err, ok := s.errorMap[id]; ok && err != nil {
		return err
	}

	for _, x := range s.items {
		if ptr.Equal(x.webAcl.Id, input.Id) {
			if x.lockToken != lockToken {
				return &smithy.GenericAPIError{
					Code:    "WAFOptimisticLockException",
					Message: "The resource has been modified since you last retrieved it",
				}
			}

			// Update the WebACL with deep copies to avoid shared references
			defaultActionCopy, err := util.JsonClone(input.DefaultAction)
			if err != nil {
				return err
			}
			rulesCopy, err := util.JsonClone(input.Rules)
			if err != nil {
				return err
			}
			visibilityConfigCopy, err := util.JsonClone(input.VisibilityConfig)
			if err != nil {
				return err
			}
			customResponseBodiesCopy, err := util.JsonClone(input.CustomResponseBodies)
			if err != nil {
				return err
			}

			x.webAcl.DefaultAction = defaultActionCopy
			x.webAcl.Rules = rulesCopy
			x.webAcl.VisibilityConfig = visibilityConfigCopy
			x.webAcl.CustomResponseBodies = customResponseBodiesCopy
			x.lockToken = uuid.NewString()

			return nil
		}
	}

	return &smithy.GenericAPIError{
		Code:    "WAFNonexistentItemException",
		Message: fmt.Sprintf("WebACL with id %s does not exist", id),
	}
}

func (s *webAclStore) DeleteWebACL(ctx context.Context, name, id string, scope types.Scope, lockToken string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[id]; ok && err != nil {
		return err
	}

	deleted := false
	for _, x := range s.items {
		if ptr.Equal(x.webAcl.Id, new(id)) {
			if x.lockToken != lockToken {
				return &smithy.GenericAPIError{
					Code:    "WAFOptimisticLockException",
					Message: "The resource has been modified since you last retrieved it",
				}
			}
			deleted = true
			break
		}
	}

	if !deleted {
		return &smithy.GenericAPIError{
			Code:    "WAFNonexistentItemException",
			Message: fmt.Sprintf("WebACL with id %s does not exist", id),
		}
	}

	s.items = pie.Filter(s.items, func(x *webAclEntry) bool {
		return !ptr.Equal(x.webAcl.Id, new(id))
	})

	return nil
}

func (s *webAclStore) ListWebACLs(ctx context.Context, scope types.Scope) ([]types.WebACLSummary, error) {
	s.m.Lock()
	defer s.m.Unlock()

	summaries := pie.Map(s.items, func(e *webAclEntry) types.WebACLSummary {
		return types.WebACLSummary{
			Id:          e.webAcl.Id,
			Name:        e.webAcl.Name,
			ARN:         e.webAcl.ARN,
			Description: e.webAcl.Description,
			LockToken:   new(e.lockToken),
		}
	})

	summariesCopy, err := util.JsonClone(summaries)
	if err != nil {
		return nil, err
	}

	return summariesCopy, nil
}

func (s *webAclStore) SetWebAclError(id string, err error) {
	s.errorMap[id] = err
}

func (s *webAclStore) InitiateWebAcl(id, name string, scope types.Scope) {
	s.m.Lock()
	defer s.m.Unlock()

	arn := awsutil.Waf2Arn(s.region, s.account, name, id)
	lockToken := uuid.NewString()

	item := &webAclEntry{
		webAcl: types.WebACL{
			Id:       new(id),
			Name:     new(name),
			ARN:      new(arn),
			Capacity: 100,
		},
		lockToken: lockToken,
	}

	s.items = append(s.items, item)
}

func (s *webAclStore) WafClient() awsclient.Wafv2Client {
	return s
}
