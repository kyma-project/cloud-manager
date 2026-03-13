package mock2

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/elliotchance/pie/v2"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
# KEY
createTime: '2024-11-26T10:17:35.030507Z'
etag: mYc1234/gN43218/1234Gw==
name: tagKeys/281234912342923
namespacedName: my-project/my-tag
parent: projects/812343123418
shortName: my-tag
updateTime: '2024-11-26T10:17:35.030507Z'

---

# VALUE
createTime: '2024-11-26T10:17:48.338685Z'
etag: kiQl1234kD1234Rd1234qg==
name: tagValues/212346412347300
namespacedName: my-project/my-tag/my-value
parent: tagKeys/281234912342923
shortName: my-value
updateTime: '2024-11-26T10:17:48.338685Z'

---

# BINDING
name: tagBindings/%2F%2Fcompute.googleapis.com%2Fprojects%2F812343123418%2Fglobal%2Fnetworks%2F6812341912344012340/tagValues/212346412347300
parent: //compute.googleapis.com/projects/812343123418/global/networks/6812341912344012340
tagValue: tagValues/212346412347300
tagValueNamespacedName: my-project/my-tag/my-value

---

# EffectiveTag
namespacedTagKey: my-project/my-tag
namespacedTagValue: my-project/my-tag/my-value
tagKey: tagKeys/281234912342923
tagKeyParentName: projects/812343123418
tagValue: tagValues/212346412347300
*/

type ResourceManagerConfig interface {
	GetTagKeyByShortNameNoLock(keyShortName string) *resourcemanagerpb.TagKey
	GetTagValueByShortNameNoLock(keyFullName, valueShortName string) *resourcemanagerpb.TagValue
	GetTagValueByFullNameNoLock(valueFullName string) *resourcemanagerpb.TagValue
	GetTagBindingNoLock(bindingName string) *resourcemanagerpb.TagBinding
	TagBindingToEffectiveTagNoLock(binding *resourcemanagerpb.TagBinding) (*resourcemanagerpb.TagKey, *resourcemanagerpb.TagValue, *resourcemanagerpb.EffectiveTag, error)
}

func (s *store) GetTagKeyByShortNameNoLock(keyShortName string) *resourcemanagerpb.TagKey {
	for _, item := range s.tagKeys.items {
		if item.Obj.ShortName == keyShortName {
			return item.Obj
		}
	}
	return nil
}

func (s *store) GetTagValueByShortNameNoLock(keyFullName, valueShortName string) *resourcemanagerpb.TagValue {
	for _, item := range s.tagValues.items {
		if item.Obj.Parent == keyFullName && item.Obj.ShortName == valueShortName {
			return item.Obj
		}
	}
	return nil
}

func (s *store) GetTagValueByFullNameNoLock(valueFullName string) *resourcemanagerpb.TagValue {
	for _, item := range s.tagValues.items {
		if item.Obj.Name == valueFullName {
			return item.Obj
		}
	}
	return nil
}

func (s *store) GetTagBindingNoLock(bindingName string) *resourcemanagerpb.TagBinding {
	for _, item := range s.tagBindings.items {
		if item.Obj.Name == bindingName {
			return item.Obj
		}
	}
	return nil
}

// Keys ===========================================================================

func (s *store) GetTagKey(ctx context.Context, req *resourcemanagerpb.GetTagKeyRequest, _ ...gax.CallOption) (*resourcemanagerpb.TagKey, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("tag key name is required")
	}
	keyNd, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("tag key name is invalid: %v", err)
	}
	if keyNd.ResourceType() != gcputil.ResourceTypeTagKey {
		return nil, gcpmeta.NewBadRequestError("invalid tag key name format")
	}

	key, found := s.tagKeys.FindByName(keyNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("tag key %s not found", keyNd)
	}

	return util.Clone(key)
}

func (s *store) CreateTagKey(ctx context.Context, req *resourcemanagerpb.CreateTagKeyRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*resourcemanagerpb.TagKey], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	parentNd := gcputil.NewProjectName(s.ProjectId())
	if !parentNd.EqualString(req.TagKey.Parent) {
		return nil, gcpmeta.NewBadRequestError("parent must be %s", parentNd.String())
	}

	if strings.TrimSpace(req.TagKey.ShortName) == "" {
		return nil, gcpmeta.NewBadRequestError("short name must be specified")
	}
	if s.GetTagKeyByShortNameNoLock(req.TagKey.ShortName) != nil {
		return nil, gcpmeta.NewBadRequestError("tag key %s already exists", req.TagKey.ShortName)
	}

	key, err := util.Clone(req.TagKey)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to clone tag key: %v", common.ErrLogical, err)
	}

	keyNd := gcputil.NewTagKeyName(fmt.Sprintf("%d", rand.Uint64()))

	key.Name = keyNd.String()
	key.CreateTime = timestamppb.Now()
	key.NamespacedName = fmt.Sprintf("%s/%s", s.ProjectId(), key.ShortName)
	key.Parent = parentNd.String()

	s.tagKeys.Add(key, keyNd)

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), keyNd)
	if err := b.WithResult(key); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set create tagKey operation result: %v", common.ErrLogical, err)
	}
	if err := b.WithMetadata(&resourcemanagerpb.CreateTagKeyMetadata{}); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set create tagKey operation metadata: %v", common.ErrLogical, err)
	}
	b.WithDone(true)
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*resourcemanagerpb.TagKey](b.GetOperationPB()), nil
}

func (s *store) DeleteTagKey(ctx context.Context, req *resourcemanagerpb.DeleteTagKeyRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*resourcemanagerpb.TagKey], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	keyNd, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("tag key %s is invalid", req.Name)
	}
	if keyNd.ResourceType() != gcputil.ResourceTypeTagKey {
		return nil, gcpmeta.NewBadRequestError("tag key %s format invalid, must be tagKeys/{id}", req.Name)
	}

	key, found := s.tagKeys.FindByName(keyNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("tag key %s not found", req.Name)
	}

	// the mock will delete all key valyes and their bindings, not sure if real gcp does it or fails if any found

	// delete values
	var deletedTagValues []string
	s.tagValues = s.tagValues.FilterNotByCallback(func(item FilterableListItem[*resourcemanagerpb.TagValue]) bool {
		if keyNd.EqualString(item.Obj.Parent) {
			deletedTagValues = append(deletedTagValues, item.Obj.Name)
			return true
		}
		return false
	})

	// delete bindings
	s.tagBindings = s.tagBindings.FilterNotByCallback(func(item FilterableListItem[*resourcemanagerpb.TagBinding]) bool {
		return pie.Contains(deletedTagValues, item.Obj.TagValue)
	})

	// delete key right away since operation is returned as already resolved
	s.tagKeys = s.tagKeys.FilterNotByCallback(func(item FilterableListItem[*resourcemanagerpb.TagKey]) bool {
		return item.Name.Equal(keyNd)
	})

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), keyNd)
	if err := b.WithResult(key); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set delete tagKey operation result: %v", common.ErrLogical, err)
	}
	if err := b.WithMetadata(&resourcemanagerpb.DeleteTagKeyMetadata{}); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set delete tagKey operation metadata: %v", common.ErrLogical, err)
	}
	b.WithDone(true)
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*resourcemanagerpb.TagKey](b.GetOperationPB()), nil
}

func (s *store) ListTagKeys(ctx context.Context, req *resourcemanagerpb.ListTagKeysRequest, _ ...gax.CallOption) gcpclient.Iterator[*resourcemanagerpb.TagKey] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*resourcemanagerpb.TagKey]{
			err: ctx.Err(),
		}
	}

	if req.Parent == "" {
		return &iteratorMocked[*resourcemanagerpb.TagKey]{
			err: gcpmeta.NewBadRequestError("parent is required"),
		}
	}

	parentNd, err := gcputil.ParseNameDetail(req.Parent)
	if err != nil {
		return &iteratorMocked[*resourcemanagerpb.TagKey]{
			err: gcpmeta.NewBadRequestError("invalid parent: %v", err),
		}
	}

	list := s.tagKeys.FilterByCallback(func(item FilterableListItem[*resourcemanagerpb.TagKey]) bool {
		return parentNd.EqualString(item.Obj.Parent)
	})

	return list.ToIterator()
}

// Values ===========================================================================

func (s *store) GetTagValue(ctx context.Context, req *resourcemanagerpb.GetTagValueRequest, _ ...gax.CallOption) (*resourcemanagerpb.TagValue, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("tag value name is required")
	}
	valueNd, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("tag value name is invalid: %v", err)
	}
	if valueNd.ResourceType() != gcputil.ResourceTypeTagValue {
		return nil, gcpmeta.NewBadRequestError("invalid tag value name format")
	}

	value, found := s.tagValues.FindByName(valueNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("tag value %s not found", req.Name)
	}

	return util.Clone(value)
}

func (s *store) CreateTagValue(ctx context.Context, req *resourcemanagerpb.CreateTagValueRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*resourcemanagerpb.TagValue], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.TagValue == nil {
		return nil, gcpmeta.NewBadRequestError("tag value is required")
	}
	if strings.TrimSpace(req.TagValue.ShortName) == "" {
		return nil, gcpmeta.NewBadRequestError("tag value shortName is required")
	}
	tagKeyNd, err := gcputil.ParseNameDetail(req.TagValue.Parent)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid tag key: %v", err)
	}
	if tagKeyNd.ResourceType() != gcputil.ResourceTypeTagKey {
		return nil, gcpmeta.NewBadRequestError("tag key %s format invalid, must be tagKeys/{id}", req.TagValue.Name)
	}

	tagKey, found := s.tagKeys.FindByName(tagKeyNd)
	if !found {
		return nil, gcpmeta.NewBadRequestError("tag key %s not found", req.TagValue.Name)
	}

	// check if tagValue with that shortName already exists

	for _, item := range s.tagValues.items {
		if item.Obj.Parent == req.TagValue.Parent && item.Obj.ShortName == req.TagValue.ShortName {
			return nil, gcpmeta.NewBadRequestError("tag value %s already exists on tag key %s", req.TagValue.ShortName, tagKey.ShortName)
		}
	}

	// create tag value

	value, err := util.Clone(req.TagValue)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to clone tag value: %v", common.ErrLogical, err)
	}
	valueNd := gcputil.NewTagValueName(fmt.Sprintf("%d", rand.Uint64()))
	value.Name = valueNd.String()
	value.NamespacedName = fmt.Sprintf("%s/%s/%s", s.ProjectId(), tagKey.ShortName, value.ShortName)
	value.CreateTime = timestamppb.Now()

	s.tagValues.Add(value, valueNd)

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), valueNd)
	if err := b.WithResult(value); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set create tagValue operation result: %v", common.ErrLogical, err)
	}
	if err := b.WithMetadata(&resourcemanagerpb.CreateTagValueMetadata{}); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set create tagValue operation metadata: %v", common.ErrLogical, err)
	}
	b.WithDone(true)
	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*resourcemanagerpb.TagValue](b.GetOperationPB()), nil
}

func (s *store) DeleteTagValue(ctx context.Context, req *resourcemanagerpb.DeleteTagValueRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*resourcemanagerpb.TagValue], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("tag value name is required")
	}
	valueNd, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid tag value name: %v", err)
	}

	value, found := s.tagValues.FindByName(valueNd)
	if !found {
		return nil, gcpmeta.NewBadRequestError("tag value %s not found", req.Name)
	}

	// deleting all value bindings, not sure but assuming gcp is doing the same

	s.tagBindings = s.tagBindings.FilterNotByCallback(func(item FilterableListItem[*resourcemanagerpb.TagBinding]) bool {
		return valueNd.EqualString(item.Obj.TagValue)
	})

	s.tagValues = s.tagValues.FilterNotByCallback(func(item FilterableListItem[*resourcemanagerpb.TagValue]) bool {
		return valueNd.EqualString(item.Obj.Name)
	})

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), valueNd)
	if err := b.WithResult(value); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set delete tagValue operation result: %v", common.ErrLogical, err)
	}
	if err := b.WithMetadata(&resourcemanagerpb.DeleteTagValueMetadata{}); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set delete tagValue operation metadata: %v", common.ErrLogical, err)
	}
	b.WithDone(true)

	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*resourcemanagerpb.TagValue](b.GetOperationPB()), nil
}

func (s *store) ListTagValues(ctx context.Context, req *resourcemanagerpb.ListTagValuesRequest, _ ...gax.CallOption) gcpclient.Iterator[*resourcemanagerpb.TagValue] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*resourcemanagerpb.TagValue]{
			err: ctx.Err(),
		}
	}

	parentNd, err := gcputil.ParseNameDetail(req.Parent)
	if err != nil {
		return &iteratorMocked[*resourcemanagerpb.TagValue]{
			err: gcpmeta.NewBadRequestError("invalid tag key: %v", err),
		}
	}
	if parentNd.ResourceType() != gcputil.ResourceTypeTagKey {
		return &iteratorMocked[*resourcemanagerpb.TagValue]{
			err: gcpmeta.NewBadRequestError("tag key %s format invalid, must be tagKeys/{id}", req.Parent),
		}
	}

	list := s.tagValues.FilterByCallback(func(item FilterableListItem[*resourcemanagerpb.TagValue]) bool {
		return parentNd.EqualString(item.Obj.Parent)
	})

	return list.ToIterator()
}

// Bindings ===========================================================================

var tagBindingParentRegex = regexp.MustCompile(`^//([^/]+)/.+$`)
var validTagBindingParentServices = map[string]struct{}{
	"cloudresourcemanager.googleapis.com": {},
	"compute.googleapis.com":              {},
	"filestore.googleapis.com":            {},
	"redis.googleapis.com":                {},
}

func (s *store) CreateTagBinding(ctx context.Context, req *resourcemanagerpb.CreateTagBindingRequest, _ ...gax.CallOption) (gcpclient.ResultOperation[*resourcemanagerpb.TagBinding], error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.TagBinding == nil {
		return nil, gcpmeta.NewBadRequestError("tag binding is required")
	}
	if req.TagBinding.Parent == "" {
		return nil, gcpmeta.NewBadRequestError("tag binding parent is required")
	}
	// //cloudresourcemanager.googleapis.com/projects/123 cloudresourcemanager.googleapis.com
	// //compute.googleapis.com/projects/123/location/us-east1/instances/345 compute.googleapis.com
	match := tagBindingParentRegex.FindStringSubmatch(req.TagBinding.Parent)
	if len(match) != 2 {
		return nil, gcpmeta.NewBadRequestError("invalid tag binding parent %s, required //{service}/project/{id}/...", req.TagBinding.Parent)
	}
	if _, ok := validTagBindingParentServices[match[1]]; !ok {
		return nil, gcpmeta.NewBadRequestError("invalid tag binding parent service %s", match[1])
	}

	if req.TagBinding.TagValue == "" {
		return nil, gcpmeta.NewBadRequestError("tag binding value is required")
	}
	valueNd, err := gcputil.ParseNameDetail(req.TagBinding.TagValue)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid tag binding value %s: %v", req.TagBinding.TagValue, err)
	}
	if valueNd.ResourceType() != gcputil.ResourceTypeTagValue {
		return nil, gcpmeta.NewBadRequestError("tag binding value %s format invalid, must be tagValues/{id}", req.TagBinding.TagValue)
	}

	value, found := s.tagValues.FindByName(valueNd)
	if !found {
		return nil, gcpmeta.NewBadRequestError("tag binding value %s not found", req.TagBinding.TagValue)
	}

	binding, err := util.Clone(req.TagBinding)
	if err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to clone tag binding: %v", common.ErrLogical, err)
	}
	bindingNd := gcputil.NewTagBindingName(binding.Parent, valueNd.ResourceId())
	binding.Name = bindingNd.String()
	binding.TagValueNamespacedName = value.NamespacedName

	s.tagBindings.Add(binding, bindingNd)

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), bindingNd)
	if err := b.WithResult(binding); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set create tagBinding operation result: %v", common.ErrLogical, err)
	}
	if err := b.WithMetadata(&resourcemanagerpb.CreateTagBindingMetadata{}); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set create tagBinding operation metadata: %v", common.ErrLogical, err)
	}
	b.WithDone(true)

	s.longRunningOperations.Add(b, opName)

	return NewResultOperation[*resourcemanagerpb.TagBinding](b.GetOperationPB()), nil
}

func (s *store) DeleteTagBinding(ctx context.Context, req *resourcemanagerpb.DeleteTagBindingRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	if req.Name == "" {
		return nil, gcpmeta.NewBadRequestError("tag binding name is required")
	}
	bindingNd, err := gcputil.ParseNameDetail(req.Name)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid tag binding name %s: %v", req.Name, err)
	}
	if bindingNd.ResourceType() != gcputil.ResourceTypeTagBinding {
		return nil, gcpmeta.NewBadRequestError("tag binding name %s format invalid, must be tagBindings/{parent}/tagValues/{valueId}", req.Name)
	}

	_, found := s.tagBindings.FindByName(bindingNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("tag binding %s not found", req.Name)
	}

	s.tagBindings = s.tagBindings.FilterNotByCallback(func(item FilterableListItem[*resourcemanagerpb.TagBinding]) bool {
		return bindingNd.Equal(item.Name)
	})

	opName := s.newLongRunningOperationName()
	b := NewOperationLongRunningBuilder(opName.String(), bindingNd)
	// operation is void so no result is needed
	//if err := b.WithResult(binding); err != nil {
	//	return nil, gcpmeta.NewInternalServerError("%v: failed to set delete tagBinding operation result: %v", common.ErrLogical, err)
	//}
	if err := b.WithMetadata(&resourcemanagerpb.DeleteTagBindingMetadata{}); err != nil {
		return nil, gcpmeta.NewInternalServerError("%v: failed to set delete tagBinding operation metadata: %v", common.ErrLogical, err)
	}
	b.WithDone(true)

	s.longRunningOperations.Add(b, opName)

	return b.BuildVoidOperation(), nil
}

func (s *store) ListTagBindings(ctx context.Context, req *resourcemanagerpb.ListTagBindingsRequest, _ ...gax.CallOption) gcpclient.Iterator[*resourcemanagerpb.TagBinding] {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return &iteratorMocked[*resourcemanagerpb.TagBinding]{
			err: ctx.Err(),
		}
	}

	list := s.tagBindings.FilterByCallback(func(item FilterableListItem[*resourcemanagerpb.TagBinding]) bool {
		return item.Obj.Parent == req.Parent
	})

	return list.ToIterator()
}

func (s *store) ListEffectiveTags(ctx context.Context, req *resourcemanagerpb.ListEffectiveTagsRequest, _ ...gax.CallOption) gcpclient.Iterator[*resourcemanagerpb.EffectiveTag] {
	it := s.ListTagBindings(ctx, &resourcemanagerpb.ListTagBindingsRequest{
		Parent: req.Parent,
	}).All()

	var items []*resourcemanagerpb.EffectiveTag
	for binding, err := range it {
		if err != nil {
			return &iteratorMocked[*resourcemanagerpb.EffectiveTag]{
				err: err,
			}
		}
		_, _, effectiveTag, err := s.TagBindingToEffectiveTagNoLock(binding)
		if err != nil {
			return &iteratorMocked[*resourcemanagerpb.EffectiveTag]{
				err: err,
			}
		}

		items = append(items, effectiveTag)
	}

	return &iteratorMocked[*resourcemanagerpb.EffectiveTag]{
		items: items,
	}
}

func (s *store) TagBindingToEffectiveTagNoLock(binding *resourcemanagerpb.TagBinding) (*resourcemanagerpb.TagKey, *resourcemanagerpb.TagValue, *resourcemanagerpb.EffectiveTag, error) {
	valueNd, err := gcputil.ParseNameDetail(binding.TagValue)
	if err != nil {
		return nil, nil, nil, gcpmeta.NewInternalServerError("%v: invalid tag value %s: %v", common.ErrLogical, binding.TagValue, err)
	}
	value, found := s.tagValues.FindByName(valueNd)
	if !found {
		return nil, nil, nil, gcpmeta.NewInternalServerError("%v: tag value %s referred in binding %s not found", common.ErrLogical, binding.TagValue, binding.Name)
	}

	keyNd, err := gcputil.ParseNameDetail(value.Parent)
	if err != nil {
		return nil, value, nil, gcpmeta.NewInternalServerError("%v: invalid tag key %s: %v", common.ErrLogical, value.Parent, err)
	}
	key, found := s.tagKeys.FindByName(keyNd)
	if !found {
		return nil, value, nil, gcpmeta.NewInternalServerError("%v: tag key %s referred in value %s referred in binding %s not found", common.ErrLogical, value.Parent, value.Name, binding.Name)
	}

	effectiveTag := &resourcemanagerpb.EffectiveTag{
		TagValue:           binding.TagValue,
		NamespacedTagValue: binding.TagValueNamespacedName,
		TagKey:             value.Parent,
		NamespacedTagKey:   key.NamespacedName,
		TagKeyParentName:   key.Parent,
		Inherited:          false,
	}

	return key, value, effectiveTag, nil
}
