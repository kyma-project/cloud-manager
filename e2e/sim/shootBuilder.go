package sim

import (
	"fmt"
	"time"

	"github.com/3th1nk/cidr"
	gardeneraws "github.com/gardener/gardener-extension-provider-aws/pkg/apis/aws/v1alpha1"
	gardenerazure "github.com/gardener/gardener-extension-provider-azure/pkg/apis/azure/v1alpha1"
	gardenergcp "github.com/gardener/gardener-extension-provider-gcp/pkg/apis/gcp/v1alpha1"
	gardeneraopenstack "github.com/gardener/gardener-extension-provider-openstack/pkg/apis/openstack/v1alpha1"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardenerconstants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/hashicorp/go-multierror"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

type ShootBuilder struct {
	Obj gardenertypes.Shoot

	cpr CloudProfileRegistry

	errWithRuntime []error
}

func NewShootBuilder(cpr CloudProfileRegistry) *ShootBuilder {
	return &ShootBuilder{
		cpr: cpr,
		Obj: gardenertypes.Shoot{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenertypes.SchemeGroupVersion.String(),
				Kind:       "Shoot",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: e2econfig.Config.GardenNamespace,
			},
			Spec: gardenertypes.ShootSpec{},
		},
	}
}

func (b *ShootBuilder) WithRuntime(rt *infrastructuremanagerv1.Runtime) *ShootBuilder {
	b.errWithRuntime = nil
	cloudProfileName, ok := e2econfig.Config.CloudProfiles[rt.Spec.Shoot.Provider.Type]
	if !ok {
		b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("cloud profile for provider %q not found", rt.Spec.Shoot.Provider.Type))
		return b
	}
	profile := b.cpr.Get(cloudProfileName)
	if profile == nil {
		b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("cloud profile %q not found in the cluster", cloudProfileName))
		return b
	}
	kv := profile.GetKubernetesVersion()
	if kv == "" {
		b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("no kubernetes version found in cloud profile %q", cloudProfileName))
		return b
	}

	b.Obj.Name = rt.Spec.Shoot.Name
	b.Obj.Namespace = e2econfig.Config.GardenNamespace

	b.Obj.Labels = map[string]string{
		gardenerconstants.LabelExtensionProviderTypePrefix + rt.Spec.Shoot.Provider.Type: "true",
		"extensions.extensions.gardener.cloud/shoot-oidc-service":                        "true",
		"operatingsystemconfig.extensions.gardener.cloud/gardenlinux":                    "true",
	}

	//nolint:staticcheck
	b.Obj.Spec.CloudProfileName = ptr.To(cloudProfileName)
	b.Obj.Spec.CloudProfile = &gardenertypes.CloudProfileReference{
		Kind: "CloudProfile",
		Name: cloudProfileName,
	}

	b.Obj.Spec.Extensions = []gardenertypes.Extension{
		{
			Type: "shoot-oidc-service",
		},
	}

	b.Obj.Spec.Kubernetes.Version = ptr.Deref(rt.Spec.Shoot.Kubernetes.Version, kv)
	b.Obj.Spec.Kubernetes.KubeAPIServer = &gardenertypes.KubeAPIServerConfig{
		Requests: &gardenertypes.APIServerRequests{
			MaxNonMutatingInflight: ptr.To(int32(800)),
			MaxMutatingInflight:    ptr.To(int32(400)),
		},
		EventTTL: &metav1.Duration{Duration: time.Hour},
	}

	b.Obj.Spec.Networking = &gardenertypes.Networking{
		Type:       ptr.To(ptr.Deref(rt.Spec.Shoot.Networking.Type, "calico")),
		Nodes:      ptr.To(rt.Spec.Shoot.Networking.Nodes),
		Pods:       ptr.To(rt.Spec.Shoot.Networking.Pods),
		Services:   ptr.To(rt.Spec.Shoot.Networking.Services),
		IPFamilies: []gardenertypes.IPFamily{gardenertypes.IPFamilyIPv4},
	}

	b.Obj.Spec.Maintenance = &gardenertypes.Maintenance{
		AutoUpdate: &gardenertypes.MaintenanceAutoUpdate{
			KubernetesVersion:   true,
			MachineImageVersion: ptr.To(false),
		},
		TimeWindow: &gardenertypes.MaintenanceTimeWindow{
			Begin: "070000+0000",
			End:   "080000+0000",
		},
	}

	b.Obj.Spec.Provider.Type = rt.Spec.Shoot.Provider.Type

	controlPlaneConfig := &runtime.RawExtension{}
	infrastructureConfig := &runtime.RawExtension{}

	switch rt.Spec.Shoot.Provider.Type {
	case "gcp":
		ic := &gardenergcp.InfrastructureConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenergcp.SchemeGroupVersion.String(),
				Kind:       "InfrastructureConfig",
			},
		}
		if e2econfig.Config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.Workers = rt.Spec.Shoot.Networking.Nodes
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for GCP", e2econfig.Config.NetworkOwner))
			return b
		}
		infrastructureConfig.Object = ic

		controlPlaneConfig.Object = &gardenergcp.ControlPlaneConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenergcp.SchemeGroupVersion.String(),
				Kind:       "ControlPlaneConfig",
			},
			Zone: rt.Spec.Shoot.Provider.Workers[0].Zones[0],
		}
	case "aws":
		nodesRange := cidr.ParseNoError(rt.Spec.Shoot.Networking.Nodes)
		zoneRanges, err := nodesRange.SubNetting(cidr.MethodSubnetNum, 16)
		if err != nil {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("failed to subnet nodes CIDR %s: %w", rt.Spec.Shoot.Networking.Nodes, err))
			return b
		}
		var zones []gardeneraws.Zone
		for i, zone := range rt.Spec.Shoot.Provider.Workers[0].Zones {
			zones = append(zones, gardeneraws.Zone{
				Name:     zone,
				Internal: zoneRanges[i*3+0].CIDR().String(),
				Public:   zoneRanges[i*3+1].CIDR().String(),
				Workers:  zoneRanges[i*3+2].CIDR().String(),
			})
		}

		ic := &gardeneraws.InfrastructureConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardeneraws.SchemeGroupVersion.String(),
				Kind:       "InfrastructureConfig",
			},
			Networks: gardeneraws.Networks{
				Zones: zones,
			},
		}

		if e2econfig.Config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.VPC.CIDR = ptr.To(rt.Spec.Shoot.Networking.Nodes)
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for AWS", e2econfig.Config.NetworkOwner))
			return b
		}

		infrastructureConfig.Object = ic

		controlPlaneConfig.Object = &gardeneraws.ControlPlaneConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardeneraws.SchemeGroupVersion.String(),
				Kind:       "ControlPlaneConfig",
			},
			CloudControllerManager: &gardeneraws.CloudControllerManagerConfig{
				UseCustomRouteController: ptr.To(true),
			},
			Storage: &gardeneraws.Storage{
				ManagedDefaultClass: ptr.To(true),
			},
		}
	case "azure":
		nodesRange := cidr.ParseNoError(rt.Spec.Shoot.Networking.Nodes)
		zoneRanges, err := nodesRange.SubNetting(cidr.MethodSubnetNum, 4)
		if err != nil {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("failed to subnet nodes CIDR %s: %w", rt.Spec.Shoot.Networking.Nodes, err))
			return b
		}
		var zones []gardenerazure.Zone
		for i := range rt.Spec.Shoot.Provider.Workers[0].Zones {
			zones = append(zones, gardenerazure.Zone{
				Name: int32(i + 1),
				CIDR: zoneRanges[i].CIDR().String(),
				NatGateway: &gardenerazure.ZonedNatGatewayConfig{
					Enabled:                      true,
					IdleConnectionTimeoutMinutes: ptr.To(int32(4)),
				},
			})
		}

		ic := &gardenerazure.InfrastructureConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenerazure.SchemeGroupVersion.String(),
				Kind:       "InfrastructureConfig",
			},
			Networks: gardenerazure.NetworkConfig{
				Zones: zones,
			},
			Zoned: true,
		}

		if e2econfig.Config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.VNet.CIDR = ptr.To(rt.Spec.Shoot.Networking.Nodes)
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for Azure", e2econfig.Config.NetworkOwner))
			return b
		}

		infrastructureConfig.Object = ic

		controlPlaneConfig.Object = &gardenerazure.ControlPlaneConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenerazure.SchemeGroupVersion.String(),
				Kind:       "ControlPlaneConfig",
			},
		}
	case "openstack":
		ic := &gardeneraopenstack.InfrastructureConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardeneraopenstack.SchemeGroupVersion.String(),
				Kind:       "InfrastructureConfig",
			},
		}

		if e2econfig.Config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.Workers = rt.Spec.Shoot.Networking.Nodes
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for OpenStack", e2econfig.Config.NetworkOwner))
			return b
		}

		infrastructureConfig.Object = ic

		controlPlaneConfig.Object = &gardeneraopenstack.ControlPlaneConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardeneraopenstack.SchemeGroupVersion.String(),
				Kind:       "ControlPlaneConfig",
			},
			LoadBalancerProvider: "f5",
		}
	}

	b.Obj.Spec.Provider.ControlPlaneConfig = controlPlaneConfig
	b.Obj.Spec.Provider.InfrastructureConfig = infrastructureConfig
	b.Obj.Spec.Provider.Workers = rt.Spec.Shoot.Provider.Workers

	b.Obj.Spec.Region = rt.Spec.Shoot.Region
	b.Obj.Spec.SecretBindingName = ptr.To(rt.Spec.Shoot.SecretBindingName)

	return b
}

func (b *ShootBuilder) Validate() error {
	var result error
	if b.errWithRuntime != nil {
		result = multierror.Append(result, b.errWithRuntime...)
	}
	if b.Obj.Namespace == "" {
		result = multierror.Append(result, fmt.Errorf("namespace must be set"))
	}
	if b.Obj.Name == "" {
		result = multierror.Append(result, fmt.Errorf("name must be set"))
	}
	return result
}

func (b *ShootBuilder) Build() *gardenertypes.Shoot {
	return &b.Obj
}
