package migrateFinalizers

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type successHandler interface {
	IsRecorded(ctx context.Context) (bool, error)
	Record(ctx context.Context) error
}

// skrSuccessHandler ===================================================================================

func newSkrSuccessHandler(kymaName string, namespace string, kcpClient client.Client) successHandler {
	return &skrSuccessHandler{
		kymaName:  kymaName,
		namespace: namespace,
		kcpClient: kcpClient,
	}
}

const successAnnotation = "cloud-manager.kyma-project.io/finalizer-migration"

type skrSuccessHandler struct {
	kymaName  string
	namespace string
	kcpClient client.Client
}

func (h *skrSuccessHandler) IsRecorded(ctx context.Context) (bool, error) {
	kyma := util.NewKymaUnstructured()
	err := h.kcpClient.Get(ctx, client.ObjectKey{Namespace: h.namespace, Name: h.kymaName}, kyma)
	if err != nil {
		return false, fmt.Errorf("error loading Kyma %s in skrSuccessHandler.IsRecorded: %w", h.kymaName, err)
	}
	_, marked := kyma.GetAnnotations()[successAnnotation]
	return marked, nil
}

func (h *skrSuccessHandler) Record(ctx context.Context) error {
	kyma := util.NewKymaUnstructured()
	err := h.kcpClient.Get(ctx, client.ObjectKey{Namespace: h.namespace, Name: h.kymaName}, kyma)
	if err != nil {
		return fmt.Errorf("error loading Kyma %s in skrSuccessHandler Record: %w", h.kymaName, err)
	}
	_, err = composed.PatchObjAddAnnotation(ctx, successAnnotation, "true", kyma, h.kcpClient)
	if err != nil {
		return fmt.Errorf("error patching kyma %s in skrSuccessHandler Record: %w", h.kymaName, err)
	}
	return nil
}

// kcpSuccessHandler ===================================================================================

func newKcpSuccessHandler(namespace string, kcpClient client.Client) successHandler {
	return &kcpSuccessHandler{
		namespace: namespace,
		kcpClient: kcpClient,
	}
}

const kcpConfigMapName = "cloud-manager-finalizer-migration"

type kcpSuccessHandler struct {
	namespace string
	kcpClient client.Client
}

func (h *kcpSuccessHandler) IsRecorded(ctx context.Context) (bool, error) {
	cm := &corev1.ConfigMap{}
	err := h.kcpClient.Get(ctx, client.ObjectKey{Namespace: h.namespace, Name: kcpConfigMapName}, cm)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error getting cm-finalizer-migration configmap: %w", err)
	}
	return true, nil
}

func (h *kcpSuccessHandler) Record(ctx context.Context) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: h.namespace,
			Name:      kcpConfigMapName,
		},
		Data: map[string]string{
			"migrated": "true",
		},
	}
	err := h.kcpClient.Create(ctx, cm)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error creating cm-finalizer-migration configmap: %w", err)
	}
	return nil
}
