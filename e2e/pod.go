package e2e

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

type PodBuilder interface {
	//ConfigMap() *corev1.ConfigMap
	ExtraResources() map[string]client.Object
	ExtraResourceObjects() []client.Object
	AddExtraResource(extra ...ExtraResource) PodBuilder
	Pod() *corev1.Pod
	Create(ctx context.Context, clstr ClusterInSession) error
	Delete(ctx context.Context, clnt client.Client) error
	WithLabel(key, value string) PodBuilder
	WithAnnotation(key, value string) PodBuilder
	WithPodDetails(arr ...PodDetailFunc) PodBuilder

	DumpYaml(scheme *runtime.Scheme) ([]byte, error)
}

type podBuilder struct {
	//cm  *corev1.ConfigMap
	extraResources map[string]client.Object
	pod            *corev1.Pod
}

func NewPodBuilder(name, namespace, image string) PodBuilder {
	b := &podBuilder{
		extraResources: make(map[string]client.Object),
		pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            name,
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/bin/sh", "-c"},
						Args: []string{
							"echo 'Noop! I am not given any script to run.'",
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}

	return b
}

func (b *podBuilder) ExtraResources() map[string]client.Object {
	return b.extraResources
}

func (b *podBuilder) ExtraResourceObjects() []client.Object {
	result := make([]client.Object, 0, len(b.extraResources))
	for _, res := range b.extraResources {
		result = append(result, res)
	}
	return result
}

func (b *podBuilder) Pod() *corev1.Pod {
	return b.pod
}

func (b *podBuilder) AddExtraResource(extra ...ExtraResource) PodBuilder {
	for _, res := range extra {
		if res.Key != "" && res.Obj != nil {
			b.extraResources[res.Key] = res.Obj
		}
	}
	return b
}

func (b *podBuilder) Create(ctx context.Context, clstr ClusterInSession) error {
	for name, res := range b.extraResources {
		if err := clstr.GetClient().Create(ctx, res); err != nil {
			return fmt.Errorf("failed to create extra resource %s: %w", name, err)
		}
		clstr.DeleteOnTerminate(res)
	}
	if err := clstr.GetClient().Create(ctx, b.pod); err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}
	clstr.DeleteOnTerminate(b.pod)
	return nil
}

func (b *podBuilder) Delete(ctx context.Context, clnt client.Client) error {
	for name, res := range b.extraResources {
		if err := clnt.Delete(ctx, res); err != nil {
			return fmt.Errorf("failed to delete extra resource %s: %w", name, err)
		}
	}
	if err := clnt.Delete(ctx, b.pod); err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	return nil
}

func (b *podBuilder) WithImage(v string) PodBuilder {
	b.pod.Spec.Containers[0].Image = v
	return b
}

func (b *podBuilder) WithLabel(key, value string) PodBuilder {
	if b.pod.Labels == nil {
		b.pod.Labels = make(map[string]string)
	}
	b.pod.Labels[key] = value
	return b
}

func (b *podBuilder) WithAnnotation(key, value string) PodBuilder {
	if b.pod.Annotations == nil {
		b.pod.Annotations = make(map[string]string)
	}
	b.pod.Annotations[key] = value
	return b
}

func (b *podBuilder) DumpYaml(scheme *runtime.Scheme) ([]byte, error) {
	var result []byte

	if b.pod.APIVersion == "" {
		b.pod.APIVersion = "v1"
	}
	if b.pod.Kind == "" {
		b.pod.Kind = "Pod"
	}

	podYaml, err := yaml.Marshal(b.pod)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pod yaml: %w", err)
	}
	result = append(result, podYaml...)

	for key, obj := range b.extraResources {
		if obj.GetObjectKind().GroupVersionKind().Kind == "" || obj.GetObjectKind().GroupVersionKind().Group == "" {
			gvk, err := apiutil.GVKForObject(obj, scheme)
			if err != nil {
				return nil, fmt.Errorf("failed to determine GVK for object %T with name %s: %w", obj, obj.GetName(), err)
			}
			obj.GetObjectKind().SetGroupVersionKind(gvk)
		}
		objYaml, err := yaml.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal extra resource %q yaml: %w", key, err)
		}
		result = append(result, []byte("\n---\n")...)
		result = append(result, objYaml...)
	}

	return result, nil
}

type ExtraResource struct {
	Key string
	Obj client.Object
}

type PodDetailFunc func(bb PodBuilder)

func (b *podBuilder) WithPodDetails(arr ...PodDetailFunc) PodBuilder {
	for _, fn := range arr {
		fn(b)
	}
	return b
}

// pod details ================================

func PodWithImage(image string) PodDetailFunc {
	return func(bb PodBuilder) {
		bb.Pod().Spec.Containers[0].Image = image
	}
}

func PodWithScript(scriptLines []string) PodDetailFunc {
	scriptTemplate := `#!/bin/bash
set -e
%s
`
	return func(bb PodBuilder) {
		name := bb.Pod().Name
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: bb.Pod().Namespace,
				Name:      name,
			},
			Data: map[string]string{
				fmt.Sprintf("%s.sh", name): fmt.Sprintf(scriptTemplate, strings.Join(scriptLines, "\n")),
			},
		}
		bb.AddExtraResource(ExtraResource{Key: "script", Obj: cm})
		bb.WithPodDetails(PodWithMountFromConfigMap(name, "", "/script/"+name, ptr.To(int32(0755))))
		bb.Pod().Spec.Containers[0].Args = []string{
			fmt.Sprintf("/script/%s/%s.sh", name, name),
		}
	}
}

func PodWithFixedEnvVars(envVars map[string]string) PodDetailFunc {
	return func(bb PodBuilder) {
		for k, v := range envVars {
			bb.Pod().Spec.Containers[0].Env = append(bb.Pod().Spec.Containers[0].Env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}
}

func PodWithFixedEnvVar(name, value string) PodDetailFunc {
	return func(bb PodBuilder) {
		bb.Pod().Spec.Containers[0].Env = append(bb.Pod().Spec.Containers[0].Env, corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}
}

func PodWithEnvFromSecret(envVarName string, secretName string, key string) PodDetailFunc {
	return func(bb PodBuilder) {
		bb.Pod().Spec.Containers[0].Env = append(bb.Pod().Spec.Containers[0].Env, corev1.EnvVar{
			Name: envVarName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: key,
				},
			},
		})
	}
}

func PodWithEnvFromConfigMap(envVarName string, configMapName string, key string) PodDetailFunc {
	return func(bb PodBuilder) {
		bb.Pod().Spec.Containers[0].Env = append(bb.Pod().Spec.Containers[0].Env, corev1.EnvVar{
			Name: envVarName,
			ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					Key: key,
				},
			},
		})
	}
}

func PodWithMountFromConfigMap(configMapName string, volumeName string, mountPath string, defaultMode *int32) PodDetailFunc {
	if volumeName == "" {
		volumeName = configMapName
	}
	if mountPath == "" {
		mountPath = "/mnt/" + volumeName
	}
	mountPath = strings.TrimSuffix(mountPath, "/")
	if defaultMode == nil {
		defaultMode = ptr.To(int32(0644))
	}
	return func(bb PodBuilder) {
		bb.Pod().Spec.Volumes = append(bb.Pod().Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					DefaultMode: defaultMode,
				},
			},
		})
		bb.Pod().Spec.Containers[0].VolumeMounts = append(bb.Pod().Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}
}

func PodWithMountFromSecret(secretName string, volumeName string, mountPath string) PodDetailFunc {
	if volumeName == "" {
		volumeName = secretName
	}
	if mountPath == "" {
		mountPath = "/mnt/" + volumeName
	}
	return func(bb PodBuilder) {
		bb.Pod().Spec.Volumes = append(bb.Pod().Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})
		bb.Pod().Spec.Containers[0].VolumeMounts = append(bb.Pod().Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}
}

func PodWithMountFromPVC(pvcName string, volumeName string, mountPath string) PodDetailFunc {
	if volumeName == "" {
		volumeName = pvcName
	}
	if mountPath == "" {
		mountPath = "/mnt" + volumeName
	}
	return func(bb PodBuilder) {
		bb.Pod().Spec.Volumes = append(bb.Pod().Spec.Volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		})
		bb.Pod().Spec.Containers[0].VolumeMounts = append(bb.Pod().Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}
}

func PodWithArguments(args ...string) PodDetailFunc {
	return func(bb PodBuilder) {
		bb.Pod().Spec.Containers[0].Args = args
	}
}
