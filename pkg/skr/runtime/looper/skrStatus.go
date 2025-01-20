package looper

import (
	"context"
	"fmt"
	"github.com/elliotchance/pie/v2"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// SkrStatusRepo ============================================================================

func NewSkrStatusRepo(kcpClient client.Client) SkrStatusRepo {
	return &skrStatusRepo{
		kcpClient: kcpClient,
	}
}

type SkrStatusRepo interface {
	Load(ctx context.Context, name, namespace string) (*cloudcontrol1beta1.SkrStatus, error)
	Save(ctx context.Context, skrStatus *cloudcontrol1beta1.SkrStatus) error
}

type skrStatusRepo struct {
	kcpClient client.Client
}

func (r *skrStatusRepo) Load(ctx context.Context, name, namespace string) (*cloudcontrol1beta1.SkrStatus, error) {
	skrStatus := &cloudcontrol1beta1.SkrStatus{}
	err := r.kcpClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, skrStatus)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error loading SkrStatus %s/%s: %w", namespace, name, err)
	}
	if apierrors.IsNotFound(err) {
		skrStatus.Name = name
		skrStatus.Namespace = namespace
		skrStatus.TypeMeta.Kind = "SkrStatus"
		skrStatus.TypeMeta.APIVersion = cloudcontrol1beta1.GroupVersion.String()
		return skrStatus, nil
	}
	return skrStatus, nil
}

func (r *skrStatusRepo) Save(ctx context.Context, skrStatus *cloudcontrol1beta1.SkrStatus) error {
	return r.kcpClient.Patch(ctx, skrStatus, client.Apply, client.ForceOwnership, client.FieldOwner(common.FieldOwner))
}

// SkrStatusSaver ==========================================================================

func NewSkrStatusSaver(repo SkrStatusRepo, namespace string) SkrStatusSaver {
	return &skrStatusSaver{
		repo:      repo,
		namespace: namespace,
	}
}

type SkrStatusSaver interface {
	Save(ctx context.Context, skrStatus *SkrStatus) error
}

type skrStatusSaver struct {
	repo      SkrStatusRepo
	namespace string
}

func (s *skrStatusSaver) Save(ctx context.Context, skrStatus *SkrStatus) error {
	api, err := s.repo.Load(ctx, skrStatus.kyma, s.namespace)
	if err != nil {
		return fmt.Errorf("error loading SkrStatus: %w", err)
	}

	if api.Labels == nil {
		api.Labels = map[string]string{}
	}
	if len(skrStatus.globalAccount) > 0 {
		api.Labels[cloudcontrol1beta1.LabelScopeGlobalAccountId] = skrStatus.globalAccount
	}
	if len(skrStatus.subAccount) > 0 {
		api.Labels[cloudcontrol1beta1.LabelScopeSubaccountId] = skrStatus.subAccount
	}
	if len(skrStatus.shoot) > 0 {
		api.Labels[cloudcontrol1beta1.LabelScopeShootName] = skrStatus.shoot
	}
	if len(skrStatus.region) > 0 {
		api.Labels[cloudcontrol1beta1.LabelScopeRegion] = skrStatus.region
	}
	if len(skrStatus.brokerPlan) > 0 {
		api.Labels[cloudcontrol1beta1.LabelScopeBrokerPlanName] = skrStatus.brokerPlan
	}

	api.Spec.Kyma = skrStatus.kyma
	api.Spec.Provider = skrStatus.provider
	api.Spec.BrokerPlan = skrStatus.brokerPlan
	api.Spec.GlobalAccount = skrStatus.globalAccount
	api.Spec.SubAccount = skrStatus.subAccount
	api.Spec.Region = skrStatus.region
	api.Spec.Shoot = skrStatus.shoot

	api.Spec.PastConnections = append(api.Spec.PastConnections, metav1.Now())
	if len(api.Spec.PastConnections) > 10 {
		api.Spec.PastConnections = api.Spec.PastConnections[len(api.Spec.PastConnections)-10:]
	}

	sum := float64(0)
	count := 0
	last := (*time.Time)(nil)
	for _, dt := range api.Spec.PastConnections {
		if last != nil {
			dur := dt.Sub(*last)
			sum += dur.Seconds()
			count++
		}
		last = ptr.To(dt.Time)
	}
	api.Spec.AverageIntervalSeconds = 0
	if count > 0 {
		api.Spec.AverageIntervalSeconds = int(math.Round(sum / float64(count)))
	}

	api.Spec.Conditions = pie.Map(skrStatus.handles, func(x *KindHandle) cloudcontrol1beta1.SkrStatusCondition {
		return cloudcontrol1beta1.SkrStatusCondition{
			Title:           x.title,
			ObjKindGroup:    x.objKindGroup,
			CrdKindGroup:    x.crdKindGroup,
			BusolaKindGroup: x.busolaKindGroup,
			Feature:         x.feature,
			ObjName:         x.objName,
			ObjNamespace:    x.objNamespace,
			Filename:        x.filename,
			Ok:              x.ok,
			Outcomes:        x.outcomes,
		}
	})

	return s.repo.Save(ctx, api.CloneForPatch())
}

// SkrStatus =============================================================================

func NewSkrStatus(ctx context.Context) *SkrStatus {
	reader := feature.NewContextReaderFromCtx(ctx)
	return &SkrStatus{
		kyma:          reader.Kyma(),
		provider:      reader.Provider(),
		brokerPlan:    reader.BrokerPlan(),
		globalAccount: reader.GlobalAccount(),
		subAccount:    reader.SubAccount(),
		region:        reader.Region(),
		shoot:         reader.Shoot(),

		ok: false,
	}
}

type SkrStatus struct {
	IsSaved bool

	kyma          string
	provider      string
	brokerPlan    string
	globalAccount string
	subAccount    string
	region        string
	shoot         string

	handles []*KindHandle

	ok      bool
	outcome string
}

type KindHandle struct {
	title           string
	objKindGroup    string
	crdKindGroup    string
	busolaKindGroup string
	feature         string
	objName         string
	objNamespace    string
	filename        string
	ok              bool
	outcomes        []string
}

// SkrStatus ==================================================================

// NotReady called when checker determines that SKR is not ready
func (s *SkrStatus) NotReady() {
	s.outcome = "SkrIsNotReady"
}

func (s *SkrStatus) Connected() {
	s.outcome = "Connected"
	s.ok = true
}

// Handle called for each manifest found in the installation files. The outcome is recorded by called a method on the returned handle
func (s *SkrStatus) Handle(ctx context.Context, title string) *KindHandle {
	reader := feature.NewContextReaderFromCtx(ctx)
	h := &KindHandle{
		title:           title,
		objKindGroup:    reader.ObjKindGroup(),
		crdKindGroup:    reader.CrdKindGroup(),
		busolaKindGroup: reader.BusolaKindGroup(),
		feature:         reader.Feature(),
	}
	s.handles = append(s.handles, h)
	return h
}

// KindHandle =================================================

func (h *KindHandle) WithObj(obj client.Object) {
	h.objName = obj.GetName()
	h.objNamespace = obj.GetNamespace()
}

func (h *KindHandle) WithFilename(filename string) {
	h.filename = filename
}

// NotSupportedByProvider called if manifest does not support the SKR provider
func (h *KindHandle) NotSupportedByProvider() {
	h.outcomes = append(h.outcomes, "NotSupportedByProvider")
}

// AlreadyExistsWithSameVersion called if manifest already exists in SKR with the version equal to desired in the installation file
func (h *KindHandle) AlreadyExistsWithSameVersion(v string) {
	h.outcomes = append(h.outcomes, "AlreadyExistsWithSameVersion", fmt.Sprintf("Version: %s", v))
	h.ok = true
}

// Updating called if manifest already exists in SKR but with different version then the one in the installation file
func (h *KindHandle) Updating(existingVersion string, desiredVersion string) {
	h.outcomes = append(h.outcomes, "Updating", fmt.Sprintf("ExistingVersion: %s", existingVersion), fmt.Sprintf("DesiredVersion: %s", desiredVersion))
}

// ApiDisabled called if manifest does not exist in SKR but its feature is disabled
func (h *KindHandle) ApiDisabled() {
	h.outcomes = append(h.outcomes, "ApiDisabled")
}

// Creating called if manifest does not exist in SKR and its feature is enabled
func (h *KindHandle) Creating() {
	h.outcomes = append(h.outcomes, "Creating")
}

// Starting called if indexer or controller is started
func (h *KindHandle) Starting() {
	h.outcomes = append(h.outcomes, "Starting")
}

// Error called when creating, updating or starting got an error
func (h *KindHandle) Error(err error) {
	h.outcomes = append(h.outcomes, fmt.Sprintf("Error: %v", err))
}

// SpecCopyError called if copy spec for update got an error. This is likely a developer logical error
// and means that the code must be changed to do copy properly.
func (h *KindHandle) SpecCopyError(err error) {
	h.outcomes = append(h.outcomes, fmt.Sprintf("SpecCopyError: %v", err))
}

func (h *KindHandle) Success() {
	h.ok = true
}
