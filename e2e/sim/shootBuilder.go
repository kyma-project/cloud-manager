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
	e2elib "github.com/kyma-project/cloud-manager/e2e/lib"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

type ShootBuilder struct {
	obj gardenertypes.Shoot

	cpr    e2elib.CloudProfileRegistry
	config *e2econfig.ConfigType

	errWithRuntime []error
}

func NewShootBuilder(cpr e2elib.CloudProfileRegistry, config *e2econfig.ConfigType) *ShootBuilder {
	b := &ShootBuilder{
		cpr:    cpr,
		config: config,
		obj: gardenertypes.Shoot{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gardenertypes.SchemeGroupVersion.String(),
				Kind:       "Shoot",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   config.GardenNamespace,
				Annotations: config.ShootAnnotations,
			},
		},
	}
	loc, err := time.LoadLocation("Europe/Belgrade")
	if err == nil {
		if time.Now().In(loc).Hour() < 18 {
			b.obj.Spec.Hibernation = &gardenertypes.Hibernation{
				Schedules: []gardenertypes.HibernationSchedule{
					{
						Start:    ptr.To("00 20 * * 1,2,3,4,5,6,0"),
						Location: ptr.To("Europe/Belgrade"),
					},
				},
			}
		}
	}
	return b
}

func (b *ShootBuilder) WithRuntime(rt *infrastructuremanagerv1.Runtime) *ShootBuilder {
	b.errWithRuntime = nil
	cloudProfileName, ok := b.config.CloudProfiles[rt.Spec.Shoot.Provider.Type]
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

	b.obj.Name = rt.Spec.Shoot.Name
	b.obj.Namespace = b.config.GardenNamespace

	b.obj.Labels = map[string]string{
		gardenerconstants.LabelExtensionProviderTypePrefix + rt.Spec.Shoot.Provider.Type: "true",
		"extensions.extensions.gardener.cloud/shoot-oidc-service":                        "true",
		"operatingsystemconfig.extensions.gardener.cloud/gardenlinux":                    "true",
	}

	//nolint:staticcheck
	//b.obj.Spec.CloudProfileName = ptr.To(cloudProfileName)
	b.obj.Spec.CloudProfile = &gardenertypes.CloudProfileReference{
		Kind: "CloudProfile",
		Name: cloudProfileName,
	}

	b.obj.Spec.Extensions = []gardenertypes.Extension{
		{
			Type: "shoot-oidc-service",
		},
	}

	b.obj.Spec.Kubernetes.Version = ptr.Deref(rt.Spec.Shoot.Kubernetes.Version, kv)
	b.obj.Spec.Kubernetes.KubeAPIServer = &gardenertypes.KubeAPIServerConfig{
		Requests: &gardenertypes.APIServerRequests{
			MaxNonMutatingInflight: ptr.To(int32(800)),
			MaxMutatingInflight:    ptr.To(int32(400)),
		},
		EventTTL: &metav1.Duration{Duration: time.Hour},
	}

	b.obj.Spec.Networking = &gardenertypes.Networking{
		Type:       ptr.To(ptr.Deref(rt.Spec.Shoot.Networking.Type, "calico")),
		Nodes:      ptr.To(rt.Spec.Shoot.Networking.Nodes),
		Pods:       ptr.To(rt.Spec.Shoot.Networking.Pods),
		Services:   ptr.To(rt.Spec.Shoot.Networking.Services),
		IPFamilies: []gardenertypes.IPFamily{gardenertypes.IPFamilyIPv4},
	}

	b.obj.Spec.Maintenance = &gardenertypes.Maintenance{
		AutoUpdate: &gardenertypes.MaintenanceAutoUpdate{
			KubernetesVersion:   true,
			MachineImageVersion: ptr.To(false),
		},
		TimeWindow: &gardenertypes.MaintenanceTimeWindow{
			Begin: "070000+0000",
			End:   "080000+0000",
		},
	}

	b.obj.Spec.Provider.Type = rt.Spec.Shoot.Provider.Type

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
		if b.config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.Workers = rt.Spec.Shoot.Networking.Nodes
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for GCP", b.config.NetworkOwner))
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

		if b.config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.VPC.CIDR = ptr.To(rt.Spec.Shoot.Networking.Nodes)
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for AWS", b.config.NetworkOwner))
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
			LoadBalancerController: &gardeneraws.LoadBalancerControllerConfig{
				Enabled: true,
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

		if b.config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.VNet.CIDR = ptr.To(rt.Spec.Shoot.Networking.Nodes)
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for Azure", b.config.NetworkOwner))
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
			FloatingPoolName: "FloatingIP-external-kyma-01",
		}

		if b.config.NetworkOwner == e2econfig.NetworkOwnerGardener {
			ic.Networks.Workers = rt.Spec.Shoot.Networking.Nodes
		} else {
			b.errWithRuntime = append(b.errWithRuntime, fmt.Errorf("network owner %q is not supported for OpenStack", b.config.NetworkOwner))
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

	b.obj.Spec.Provider.ControlPlaneConfig = controlPlaneConfig
	b.obj.Spec.Provider.InfrastructureConfig = infrastructureConfig
	b.obj.Spec.Provider.Workers = rt.Spec.Shoot.Provider.Workers

	b.obj.Spec.Region = rt.Spec.Shoot.Region
	b.obj.Spec.CredentialsBindingName = ptr.To(rt.Spec.Shoot.SecretBindingName)

	return b
}

func (b *ShootBuilder) WithAnnotations(annotations map[string]string) *ShootBuilder {
	if len(annotations) > 0 {
		if b.obj.Annotations == nil {
			b.obj.Annotations = make(map[string]string)
		}
		for k, v := range annotations {
			b.obj.Annotations[k] = v
		}
	}
	return b
}

func (b *ShootBuilder) Validate() error {
	var result error
	if b.errWithRuntime != nil {
		result = multierror.Append(result, b.errWithRuntime...)
	}
	if b.obj.Namespace == "" {
		result = multierror.Append(result, fmt.Errorf("namespace must be set"))
	}
	if b.obj.Name == "" {
		result = multierror.Append(result, fmt.Errorf("name must be set"))
	}
	return result
}

func (b *ShootBuilder) Build() *gardenertypes.Shoot {
	return &b.obj
}
