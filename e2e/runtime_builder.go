package e2e

import (
	"errors"
	"fmt"
	"strings"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/external/keb"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

const gardenLinuxVersion = "1592.8.0"

type RuntimeBuilder struct {
	Obj keb.Runtime

	errProvider error
}

func NewRuntimeBuilder() *RuntimeBuilder {
	globalAccountId := uuid.NewString()
	subAccountId := uuid.NewString()
	name := uuid.NewString()
	shootName := util.RandomId(7)
	return &RuntimeBuilder{
		Obj: keb.Runtime{
			TypeMeta: metav1.TypeMeta{
				APIVersion: keb.GroupVersion.String(),
				Kind:       "Runtime",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: Config.KcpNamespace,
				Labels: map[string]string{
					cloudcontrolv1beta1.LabelScopeGlobalAccountId: globalAccountId,
					cloudcontrolv1beta1.LabelScopeSubaccountId:    subAccountId,
					cloudcontrolv1beta1.LabelScopeShootName:       shootName,
					cloudcontrolv1beta1.LabelKymaName:             name,
					cloudcontrolv1beta1.LabelScopeBrokerPlanName:  "", // required!!!
				},
			},
			Spec: keb.RuntimeSpec{
				Security: keb.Security{
					Administrators: Config.Administrators,
					Networking: keb.NetworkingSecurity{
						Filter: keb.Filter{
							Egress: keb.Egress{
								Enabled: true,
							},
							Ingress: &keb.Ingress{
								Enabled: false,
							},
						},
					},
				},
				Shoot: keb.RuntimeShoot{
					Kubernetes: keb.Kubernetes{
						Version: ptr.To("1.32"),
						KubeAPIServer: keb.APIServer{
							AdditionalOidcConfig: &[]keb.OIDCConfig{
								{
									OIDCConfig: gardenertypes.OIDCConfig{
										ClientID:       ptr.To(Config.OidcClientId),
										IssuerURL:      ptr.To(Config.OidcIssuerUrl),
										GroupsClaim:    ptr.To("groups"),
										GroupsPrefix:   ptr.To("-"),
										UsernameClaim:  ptr.To("sub"),
										UsernamePrefix: ptr.To("-"),
										SigningAlgs:    []string{"RS256"},
									},
								},
							},
						},
					},
					Name: shootName,
					Networking: keb.Networking{
						Nodes:    "10.250.0.0/16",
						Pods:     "10.96.0.0/13",
						Services: "10.104.0.0/13",
					},
					PlatformRegion: "cf-us10-staging",
					Provider: keb.Provider{
						Type:    "",  // required!!!
						Workers: nil, // required!!!
					},
					Purpose:           "test",
					Region:            "", // required!!! zones in workers must match this region
					SecretBindingName: "", // required!!!
				},
			},
		},
	}
}

func (b *RuntimeBuilder) WithName(name string) *RuntimeBuilder {
	b.Obj.Name = name
	return b
}

func (b *RuntimeBuilder) WithNamespace(ns string) *RuntimeBuilder {
	b.Obj.Namespace = ns
	return b
}

func (b *RuntimeBuilder) WithProvider(provider cloudcontrolv1beta1.ProviderType, region string) *RuntimeBuilder {
	b.errProvider = nil
	lProvider := strings.ToLower(string(provider))
	uProvider := strings.ToUpper(string(provider))
	b.Obj.Labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName] = lProvider
	b.Obj.Labels[cloudcontrolv1beta1.LabelScopeProvider] = uProvider

	zones, ok := providerRegions[provider][region]
	if !ok {
		b.errProvider = fmt.Errorf("unsupported region %q for provider %q", region, provider)
		return b
	}
	if len(zones) > 3 {
		zones = zones[:3] // use only first 3 zones
	}
	if len(zones) < 3 {
		b.errProvider = fmt.Errorf("too few zones for provider %q in region %q: %w", provider, region, common.ErrLogical)
		return b
	}

	var vol *gardenertypes.Volume
	volType, ok := volumeTypes[provider]
	if ok {
		vol = &gardenertypes.Volume{
			VolumeSize: "80Gi",
			Type:       ptr.To(volType),
		}
	}

	b.Obj.Spec.Shoot.Region = region
	b.Obj.Spec.Shoot.Provider.Type = lProvider
	b.Obj.Spec.Shoot.Provider.Workers = []gardenertypes.Worker{
		{
			Name: "cpu-worker-0",
			Machine: gardenertypes.Machine{
				Image: &gardenertypes.ShootMachineImage{
					Name:    "gardenlinux",
					Version: ptr.To(gardenLinuxVersion),
				},
				Type: machineTypes[provider][0],
			},
			MaxSurge:       ptr.To(intstr.FromInt32(3)),
			MaxUnavailable: ptr.To(intstr.FromInt32(0)),
			Minimum:        3,
			Maximum:        20,
			Volume:         vol,
			Zones:          zones,
		},
	}
	return b
}

func (b *RuntimeBuilder) WithAdministrators(admins ...string) *RuntimeBuilder {
	b.Obj.Spec.Security.Administrators = append(b.Obj.Spec.Security.Administrators, admins...)
	return b
}

func (b *RuntimeBuilder) WithOidc(clientId, issuerUrl string) *RuntimeBuilder {
	var data []keb.OIDCConfig
	if b.Obj.Spec.Shoot.Kubernetes.KubeAPIServer.AdditionalOidcConfig != nil {
		data = *b.Obj.Spec.Shoot.Kubernetes.KubeAPIServer.AdditionalOidcConfig
	}
	data = append(
		data,
		keb.OIDCConfig{
			OIDCConfig: gardenertypes.OIDCConfig{
				ClientID:       ptr.To(clientId),
				IssuerURL:      ptr.To(issuerUrl),
				GroupsClaim:    ptr.To("groups"),
				GroupsPrefix:   ptr.To("-"),
				UsernameClaim:  ptr.To("sub"),
				UsernamePrefix: ptr.To("-"),
				SigningAlgs:    []string{"RS256"},
			},
		},
	)
	b.Obj.Spec.Shoot.Kubernetes.KubeAPIServer.AdditionalOidcConfig = &data
	return b
}

func (b *RuntimeBuilder) WithSecretBindingName(secretBindingName string) *RuntimeBuilder {
	b.Obj.Spec.Shoot.SecretBindingName = secretBindingName
	return b
}

func (b *RuntimeBuilder) Validate() error {
	var err error
	if b.errProvider != nil {
		err = multierror.Append(err, b.errProvider)
	}
	if b.Obj.Namespace == "" {
		err = multierror.Append(err, errors.New("namespace is required"))
	}
	if b.Obj.Labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName] == "" {
		err = multierror.Append(err, fmt.Errorf("label %s is required, maybe WithProvider was not called", cloudcontrolv1beta1.LabelScopeBrokerPlanName))
	}
	if len(b.Obj.Spec.Security.Administrators) == 0 {
		err = multierror.Append(err, errors.New("at least one administrator is required"))
	}
	if b.Obj.Spec.Shoot.Kubernetes.KubeAPIServer.AdditionalOidcConfig == nil || len(*b.Obj.Spec.Shoot.Kubernetes.KubeAPIServer.AdditionalOidcConfig) == 0 {
		err = multierror.Append(err, errors.New("at least one OIDC config is required"))
	} else {
		for i, oidc := range *b.Obj.Spec.Shoot.Kubernetes.KubeAPIServer.AdditionalOidcConfig {
			if ptr.Deref(oidc.OIDCConfig.ClientID, "") == "" {
				err = multierror.Append(err, fmt.Errorf("oidc %d: client ID is required", i))
			}
			if ptr.Deref(oidc.OIDCConfig.IssuerURL, "") == "" {
				err = multierror.Append(err, fmt.Errorf("oidc %d: issuer URL is required", i))
			}
		}
	}
	if b.Obj.Spec.Shoot.Provider.Type == "" {
		err = multierror.Append(err, fmt.Errorf("provider type is required, maybe WithProvider was not called"))
	}
	if len(b.Obj.Spec.Shoot.Provider.Workers) == 0 {
		err = multierror.Append(err, errors.New("at least one worker is required, maybe WithProvider was not called"))
	}
	if b.Obj.Spec.Shoot.Region == "" {
		err = multierror.Append(err, fmt.Errorf("region is required, maybe WithProvider was not called"))
	}
	if b.Obj.Spec.Shoot.SecretBindingName == "" {
		err = multierror.Append(err, fmt.Errorf("secret binding name is required"))
	}
	return err
}

func (b *RuntimeBuilder) Build() *keb.Runtime {
	return &b.Obj
}

var providerRegions = map[cloudcontrolv1beta1.ProviderType]map[string][]string{

	// aws ec2 describe-availability-zones --region us-east-2 | jq '.AvailabilityZones[] | select(.OptInStatus == "opt-in-not-required") | .ZoneName'
	cloudcontrolv1beta1.ProviderAws: {
		"us-east-1": {
			"us-east-1a",
			"us-east-1b",
			"us-east-1c",
			//"us-east-1d",
			//"us-east-1e",
			//"us-east-1f",
		},
		"us-east-2": {
			"us-east-2a",
			"us-east-2b",
			"us-east-2c",
		},
		//"us-west-1": {
		//	"us-west-1a",
		//	"us-west-1b",
		//},
		"us-west-2": {
			"us-west-2a",
			"us-west-2b",
			"us-west-2c",
			//"us-west-2d",
		},

		"eu-central-1": {
			"eu-central-1a",
			"eu-central-1b",
			"eu-central-1c",
		},
		"eu-west-1": {
			"eu-west-1a",
			"eu-west-1b",
			"eu-west-1c",
		},
		"eu-west-2": {
			"eu-west-2a",
			"eu-west-2b",
			"eu-west-2c",
		},
		"eu-west-3": {
			"eu-west-3a",
			"eu-west-3b",
			"eu-west-3c",
		},
		"eu-north-1": {
			"eu-north-1a",
			"eu-north-1b",
			"eu-north-1c",
		},
	},

	// https://cloud.google.com/compute/docs/regions-zones
	cloudcontrolv1beta1.ProviderGCP: {
		// South Carolina
		"us-east1": {
			"us-east1-b",
			"us-east1-c",
			"us-east1-d",
		},
		// Iowa
		"us-central1": {
			"us-central1-a",
			"us-central1-b",
			"us-central1-c",
		},
		// Oregon
		"us-west1": {
			"us-west1-a",
			"us-west1-b",
			"us-west1-c",
		},

		// Hamina, Finland
		"europe-north1": {
			"europe-north1-a",
			"europe-north1-b",
			"europe-north1-c",
		},
		// St. Ghislain, Belgium
		"europe-west1": {
			"europe-west1-b",
			"europe-west1-c",
			"europe-west1-d",
		},
		// Frankfurt, Germany
		"europe-west3": {
			"europe-west3-a",
			"europe-west3-b",
			"europe-west3-c",
		},
	},

	cloudcontrolv1beta1.ProviderAzure: {
		"westeurope":  {"1", "2", "3"},
		"northeurope": {"1", "2", "3"},
		"eastus":      {"1", "2", "3"},
		"eastus2":     {"1", "2", "3"},
		"westus2":     {"1", "2", "3"},
		"westus3":     {"1", "2", "3"},
	},

	cloudcontrolv1beta1.ProviderOpenStack: {
		"eu-de-1": {
			"eu-de-1a",
			"eu-de-1b",
			"eu-de-1d",
		},
	},
}

var machineTypes = map[cloudcontrolv1beta1.ProviderType][]string{
	cloudcontrolv1beta1.ProviderAws:       {"m5.large", "m6i.large"},
	cloudcontrolv1beta1.ProviderGCP:       {"n2-standard-2"},
	cloudcontrolv1beta1.ProviderAzure:     {"Standard_D2s_v5"},
	cloudcontrolv1beta1.ProviderOpenStack: {"g_c2_m8"},
}

var volumeTypes = map[cloudcontrolv1beta1.ProviderType]string{
	cloudcontrolv1beta1.ProviderAws:   "gp3",
	cloudcontrolv1beta1.ProviderGCP:   "pd-balanced",
	cloudcontrolv1beta1.ProviderAzure: "StandardSSD_LRS",
	// nothing for ccee
	//cloudcontrolv1beta1.ProviderOpenStack: "",
}
