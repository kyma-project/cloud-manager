package migrateFinalizers

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type successHandler interface {
	IsRecorded(ctx context.Context) (bool, error)
	Record(ctx context.Context) error
}

// skrSuccessHandler ===================================================================================

func newSkrSuccessHandler(kymaName string, namespace string, kcpReader client.Reader, kcpWriter client.Writer) successHandler {
	return &skrSuccessHandler{
		kymaName:  kymaName,
		namespace: namespace,
		kcpReader: kcpReader,
		kcpWriter: kcpWriter,
	}
}

const SuccessAnnotation = "cloud-manager.kyma-project.io/finalizer-migration"

type skrSuccessHandler struct {
	kymaName  string
	namespace string
	kcpReader client.Reader
	kcpWriter client.Writer
}

func (h *skrSuccessHandler) getKymaAndRunCount(ctx context.Context) (*unstructured.Unstructured, int, error) {
	kyma := util.NewKymaUnstructured()
	err := h.kcpReader.Get(ctx, client.ObjectKey{Namespace: h.namespace, Name: h.kymaName}, kyma)
	if err != nil {
		return nil, 0, fmt.Errorf("error loading Kyma %s in skrSuccessHandler: %w", h.kymaName, err)
	}
	str, ok := kyma.GetAnnotations()[SuccessAnnotation]
	if !ok {
		return kyma, 0, nil
	}
	count, err := strconv.Atoi(str)
	if err != nil {
		return kyma, 0, nil
	}

	return kyma, count, nil
}

func (h *skrSuccessHandler) IsRecorded(ctx context.Context) (bool, error) {
	_, count, err := h.getKymaAndRunCount(ctx)
	if err != nil {
		return false, fmt.Errorf("isRecorded: %w", err)
	}
	return count >= 3, nil
}

func (h *skrSuccessHandler) Record(ctx context.Context) error {
	kyma, count, err := h.getKymaAndRunCount(ctx)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}
	count = count + 1
	_, err = composed.PatchObjMergeAnnotation(ctx, SuccessAnnotation, fmt.Sprintf("%d", count), kyma, h.kcpWriter)
	if err != nil {
		return fmt.Errorf("error patching kyma %s in skrSuccessHandler Record to %d: %w", h.kymaName, count, err)
	}
	return nil
}

// kcpSuccessHandler ===================================================================================

func newKcpSuccessHandler(namespace string, kcpReader client.Reader, kcpWriter client.Writer) successHandler {
	return &kcpSuccessHandler{
		namespace: namespace,
		kcpReader: kcpReader,
		kcpWriter: kcpWriter,
	}
}

const KcpConfigMapName = "cloud-manager-finalizer-migration"

type kcpSuccessHandler struct {
	namespace string
	kcpReader client.Reader
	kcpWriter client.Writer
}

func (h *kcpSuccessHandler) getConfigMapAndRunCount(ctx context.Context) (*corev1.ConfigMap, int, error) {
	cm := &corev1.ConfigMap{}
	cm.Name = KcpConfigMapName
	cm.Namespace = h.namespace
	err := h.kcpReader.Get(ctx, client.ObjectKeyFromObject(cm), cm)
	if apierrors.IsNotFound(err) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("error getting cm-finalizer-migration configmap: %w", err)
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	str, ok := cm.Data["runCount"]
	if !ok {
		cm.Data["runCount"] = "0"
		return cm, 0, nil
	}
	count, err := strconv.Atoi(str)
	if err != nil {
		cm.Data["runCount"] = "0"
		return cm, 0, nil
	}
	return cm, count, nil
}

func (h *kcpSuccessHandler) IsRecorded(ctx context.Context) (bool, error) {
	_, count, err := h.getConfigMapAndRunCount(ctx)
	if err != nil {
		return false, fmt.Errorf("isRecorded: %w", err)
	}
	return count >= 3, nil
}

func (h *kcpSuccessHandler) Record(ctx context.Context) error {
	cm, count, err := h.getConfigMapAndRunCount(ctx)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}
	create := false
	if cm == nil {
		create = true
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: h.namespace,
				Name:      KcpConfigMapName,
			},
			Data: map[string]string{
				"runCount": "1",
			},
		}
	} else {
		count = count + 1
		cm.Data["runCount"] = fmt.Sprintf("%d", count)
	}

	if create {
		err = h.kcpWriter.Create(ctx, cm)
	} else {
		err = h.kcpWriter.Update(ctx, cm)
	}
	if err != nil {
		txt := "updating"
		if create {
			txt = "creating"
		}
		return fmt.Errorf("error %s cm-finalizer-migration configmap: %w", txt, err)
	}
	return nil
}
