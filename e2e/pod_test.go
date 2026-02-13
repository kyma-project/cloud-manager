package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestPodBuilder(t *testing.T) {

	t.Run("with script and default mount paths", func(t *testing.T) {
		name := "pod-1"
		ns := "foo-ns-1"
		image := "busybox:1.5"
		labelKey := "e2e.cloud-manager.kyma-project.io/test-label"
		labelValue := "test-label"
		annotationKey := "e2e.cloud-manager.kyma-project.io/test-annotation"
		annotationValue := "test-annotation"
		scriptLines := []string{"echo 'this is test'"}
		pvcName := "some-pvc"
		secretName := "some-secret"
		b := NewPodBuilder(name, ns, "").
			WithLabel(labelKey, labelValue).
			WithAnnotation(annotationKey, annotationValue).
			WithPodDetails(
				PodWithScript(scriptLines),
				PodWithImage(image),
				PodWithMountFromPVC(pvcName, "", ""),
				PodWithMountFromSecret(secretName, "", ""),
			)

		cm := b.ExtraResourceObjects()[0].(*corev1.ConfigMap)
		assert.Equal(t, cm.Name, name)
		assert.Equal(t, cm.Namespace, ns)
		assert.Len(t, cm.Data, 1)
		assert.NotEmpty(t, cm.Data[fmt.Sprintf("%s.sh", name)])

		pod := b.Pod()

		assert.Equal(t, name, name)
		assert.Equal(t, ns, pod.Namespace)
		assert.Len(t, pod.Labels, 1)
		assert.Equal(t, labelValue, pod.Labels[labelKey])
		assert.Len(t, pod.Annotations, 1)
		assert.Equal(t, annotationValue, pod.Annotations[annotationKey])

		// volumes

		assert.Len(t, pod.Spec.Volumes, 3)
		assert.Equal(t, name, pod.Spec.Volumes[0].ConfigMap.Name)
		assert.Equal(t, pvcName, pod.Spec.Volumes[1].PersistentVolumeClaim.ClaimName)
		assert.Equal(t, secretName, pod.Spec.Volumes[2].Secret.SecretName)

		// containers

		assert.Len(t, pod.Spec.Containers, 1)
		assert.Equal(t, image, pod.Spec.Containers[0].Image)
		assert.Len(t, pod.Spec.Containers[0].VolumeMounts, 3)
		assert.Nil(t, pod.Spec.Containers[0].Command)

		// script volume mount

		assert.Equal(t, name, pod.Spec.Containers[0].VolumeMounts[0].Name)
		assert.Equal(t, fmt.Sprintf("/script/%s", name), pod.Spec.Containers[0].VolumeMounts[0].MountPath)

		// pvc volume mount

		assert.Equal(t, pod.Spec.Containers[0].VolumeMounts[1].Name, pvcName)
		assert.Equal(t, fmt.Sprintf("/mnt/%s", pvcName), pod.Spec.Containers[0].VolumeMounts[1].MountPath)

		// secret volume mount

		assert.Equal(t, pod.Spec.Containers[0].VolumeMounts[2].Name, secretName)
		assert.Equal(t, fmt.Sprintf("/mnt/%s", secretName), pod.Spec.Containers[0].VolumeMounts[2].MountPath)
	})
}
