package infrastructuremanagerv1

import (
	"strings"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSimpleRuntimeBuilder() *SimpleRuntimeBuilder {
	name := ""
	shootName := ""
	secretBindingName := ""
	return &SimpleRuntimeBuilder{
		Obj: &Runtime{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					cloudcontrolv1beta1.LabelRuntimeId:            name,
					cloudcontrolv1beta1.LabelScopeGlobalAccountId: "6329e93d-591f-4b1e-83ed-3dc6f9f426d7",
					cloudcontrolv1beta1.LabelScopeSubaccountId:    "f6d42db7-1195-4ff5-9787-0edb471c75cb",
					cloudcontrolv1beta1.LabelScopeShootName:       shootName,
					cloudcontrolv1beta1.LabelKymaName:             name,
					//cloudcontrolv1beta1.LabelScopeBrokerPlanName:  "aws", // required!!!
					cloudcontrolv1beta1.LabelScopeRegion: "eu-west-1",
				},
			},
			Spec: RuntimeSpec{
				Security: Security{
					Administrators: []string{"someone@sap.com"},
				},
				Shoot: RuntimeShoot{
					Name:   shootName,
					Region: "eu-west-1",
					Networking: Networking{
						Nodes: "10.250.0.0/16",
					},
					Provider: Provider{
						//Type: "aws", // required!!!
						Workers: []gardenertypes.Worker{
							{
								Name: "worker1",
								Machine: gardenertypes.Machine{
									Image: &gardenertypes.ShootMachineImage{
										Name: "gardenlinux",
									},
									Type: "m5.large",
								},
							},
						},
					},
					SecretBindingName: secretBindingName,
				},
			},
		},
	}
}

type SimpleRuntimeBuilder struct {
	Obj *Runtime
}

func (b *SimpleRuntimeBuilder) Build() *Runtime {
	if b.Obj.Spec.Shoot.Provider.Type == "" {
		panic("SimpleRuntimeBuilder - Provider must be set")
	}
	return b.Obj
}

func (b *SimpleRuntimeBuilder) WithObj(obj *Runtime) *SimpleRuntimeBuilder {
	b.Obj = obj
	return b
}

func (b *SimpleRuntimeBuilder) WithName(name string) *SimpleRuntimeBuilder {
	b.Obj.Name = name
	if b.Obj.Labels == nil {
		b.Obj.Labels = make(map[string]string)
	}
	b.Obj.Labels[cloudcontrolv1beta1.LabelRuntimeId] = name
	b.Obj.Labels[cloudcontrolv1beta1.LabelKymaName] = name
	return b
}

func (b *SimpleRuntimeBuilder) WithNamespace(namespace string) *SimpleRuntimeBuilder {
	b.Obj.Namespace = namespace
	return b
}

func (b *SimpleRuntimeBuilder) WithProvider(p cloudcontrolv1beta1.ProviderType) *SimpleRuntimeBuilder {
	lc := strings.ToLower(string(p))
	if b.Obj.Labels == nil {
		b.Obj.Labels = make(map[string]string)
	}
	b.Obj.Labels[cloudcontrolv1beta1.LabelScopeBrokerPlanName] = lc
	b.Obj.Spec.Shoot.Provider.Type = lc
	return b
}

func (b *SimpleRuntimeBuilder) WithRegion(region string) *SimpleRuntimeBuilder {
	b.Obj.Spec.Shoot.Region = region
	return b
}

func (b *SimpleRuntimeBuilder) WithBindingName(bn string) *SimpleRuntimeBuilder {
	b.Obj.Spec.Shoot.SecretBindingName = bn
	return b
}

func (b *SimpleRuntimeBuilder) WithShootName(sn string) *SimpleRuntimeBuilder {
	b.Obj.Spec.Shoot.Name = sn
	if b.Obj.Labels == nil {
		b.Obj.Labels = make(map[string]string)
	}
	b.Obj.Labels[cloudcontrolv1beta1.LabelScopeShootName] = sn
	return b
}

func (b *SimpleRuntimeBuilder) WithNodes(v string) *SimpleRuntimeBuilder {
	b.Obj.Spec.Shoot.Networking.Nodes = v
	return b
}

func (b *SimpleRuntimeBuilder) WithVpcNetworkName(v *string) *SimpleRuntimeBuilder {
	b.Obj.Spec.Shoot.Networking.VPCNetwork = v
	return b
}
