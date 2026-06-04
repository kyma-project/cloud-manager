package security

import (
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
)

func ext(name string, enabled armsecurity.IsEnabled) *armsecurity.Extension {
	return &armsecurity.Extension{Name: &name, IsEnabled: &enabled}
}

func planWithRequired(required ...ExtensionSpec) PlanSpec {
	return PlanSpec{RequiredExtensions: required}
}

func TestHasAllExtensionsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		current []*armsecurity.Extension
		plan    PlanSpec
		want    bool
	}{
		{
			name: "all required extensions are enabled",
			current: []*armsecurity.Extension{
				ext("Foo", armsecurity.IsEnabledTrue),
				ext("Bar", armsecurity.IsEnabledTrue),
			},
			plan: planWithRequired(ExtensionSpec{Name: "Foo"}, ExtensionSpec{Name: "Bar"}),
			want: true,
		},
		{
			name: "one required extension is disabled",
			current: []*armsecurity.Extension{
				ext("Foo", armsecurity.IsEnabledTrue),
				ext("Bar", armsecurity.IsEnabledFalse),
			},
			plan: planWithRequired(ExtensionSpec{Name: "Foo"}, ExtensionSpec{Name: "Bar"}),
			want: false,
		},
		{
			name: "one required extension is missing",
			current: []*armsecurity.Extension{
				ext("Foo", armsecurity.IsEnabledTrue),
			},
			plan: planWithRequired(ExtensionSpec{Name: "Foo"}, ExtensionSpec{Name: "Bar"}),
			want: false,
		},
		{
			name:    "empty required list always returns true",
			current: []*armsecurity.Extension{ext("Foo", armsecurity.IsEnabledFalse)},
			plan:    planWithRequired(),
			want:    true,
		},
		{
			name: "extra enabled extensions beyond required are ignored",
			current: []*armsecurity.Extension{
				ext("Foo", armsecurity.IsEnabledTrue),
				ext("Bar", armsecurity.IsEnabledTrue),
				ext("Baz", armsecurity.IsEnabledTrue),
			},
			plan: planWithRequired(ExtensionSpec{Name: "Foo"}),
			want: true,
		},
		{
			name:    "disable use case: nil current and empty required returns true",
			current: nil,
			plan:    planWithRequired(),
			want:    true,
		},
		{
			name:    "disable use case: nil current and non-empty required returns false",
			current: nil,
			plan:    planWithRequired(ExtensionSpec{Name: "Foo"}),
			want:    false,
		},
		{
			name: "required extension with additional properties is checked by name only",
			current: []*armsecurity.Extension{
				ext("Foo", armsecurity.IsEnabledTrue),
			},
			plan: planWithRequired(ExtensionSpec{
				Name:                          "Foo",
				AdditionalExtensionProperties: map[string]any{"SomeProp": "someValue"},
			}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.plan.hasAllExtensionsEnabled(tt.current)
			if got != tt.want {
				t.Errorf("hasAllExtensionsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlanSpec_NeedsUpdate(t *testing.T) {
	standardTier := armsecurity.PricingTierStandard
	freeTier := armsecurity.PricingTierFree
	subPlanP2 := "P2"

	plan := PlanSpec{
		PlanName: "VirtualMachines",
		SubPlan:  "P2",
		RequiredExtensions: []ExtensionSpec{
			{Name: "AgentlessVmScanning"},
		},
	}

	tests := []struct {
		name       string
		current    *armsecurity.Pricing
		targetTier armsecurity.PricingTier
		want       bool
	}{
		{
			name:       "nil current always needs update",
			current:    nil,
			targetTier: armsecurity.PricingTierStandard,
			want:       true,
		},
		{
			name: "nil properties needs update",
			current: &armsecurity.Pricing{
				Properties: nil,
			},
			targetTier: armsecurity.PricingTierStandard,
			want:       true,
		},
		{
			name: "tier mismatch needs update",
			current: &armsecurity.Pricing{
				Properties: &armsecurity.PricingProperties{
					PricingTier: &freeTier,
				},
			},
			targetTier: armsecurity.PricingTierStandard,
			want:       true,
		},
		{
			name: "free tier matches, no update needed",
			current: &armsecurity.Pricing{
				Properties: &armsecurity.PricingProperties{
					PricingTier: &freeTier,
				},
			},
			targetTier: armsecurity.PricingTierFree,
			want:       false,
		},
		{
			name: "standard tier but subplan mismatch needs update",
			current: &armsecurity.Pricing{
				Properties: &armsecurity.PricingProperties{
					PricingTier: &standardTier,
					SubPlan:     nil,
				},
			},
			targetTier: armsecurity.PricingTierStandard,
			want:       true,
		},
		{
			name: "standard tier, subplan matches, extension not enabled needs update",
			current: &armsecurity.Pricing{
				Properties: &armsecurity.PricingProperties{
					PricingTier: &standardTier,
					SubPlan:     &subPlanP2,
					Extensions: []*armsecurity.Extension{
						ext("AgentlessVmScanning", armsecurity.IsEnabledFalse),
					},
				},
			},
			targetTier: armsecurity.PricingTierStandard,
			want:       true,
		},
		{
			name: "standard tier, all matches, no update needed",
			current: &armsecurity.Pricing{
				Properties: &armsecurity.PricingProperties{
					PricingTier: &standardTier,
					SubPlan:     &subPlanP2,
					Extensions: []*armsecurity.Extension{
						ext("AgentlessVmScanning", armsecurity.IsEnabledTrue),
					},
				},
			},
			targetTier: armsecurity.PricingTierStandard,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := plan.NeedsUpdate(tt.current, tt.targetTier)
			if got != tt.want {
				t.Errorf("NeedsUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlanSpec_BuildPricing(t *testing.T) {
	plan := PlanSpec{
		PlanName: "StorageAccounts",
		SubPlan:  "DefenderForStorageV2",
		RequiredExtensions: []ExtensionSpec{
			{
				Name:                          "OnUploadMalwareScanning",
				AdditionalExtensionProperties: map[string]any{"CapGBPerMonthPerStorageAccount": "10000"},
			},
			{Name: "SensitiveDataDiscovery"},
		},
		DisabledExtensions: []string{"SomeDisabledExt"},
	}

	t.Run("free tier produces no extensions or subplan", func(t *testing.T) {
		pricing := plan.BuildPricing(armsecurity.PricingTierFree)
		if pricing.Properties == nil {
			t.Fatal("Properties must not be nil")
		}
		if *pricing.Properties.PricingTier != armsecurity.PricingTierFree {
			t.Errorf("expected Free tier, got %v", *pricing.Properties.PricingTier)
		}
		if pricing.Properties.SubPlan != nil {
			t.Errorf("expected no SubPlan for free tier, got %v", *pricing.Properties.SubPlan)
		}
		if pricing.Properties.Extensions != nil {
			t.Errorf("expected no Extensions for free tier, got %v", pricing.Properties.Extensions)
		}
	})

	t.Run("standard tier sets subplan and extensions", func(t *testing.T) {
		pricing := plan.BuildPricing(armsecurity.PricingTierStandard)
		if pricing.Properties == nil {
			t.Fatal("Properties must not be nil")
		}
		if *pricing.Properties.PricingTier != armsecurity.PricingTierStandard {
			t.Errorf("expected Standard tier, got %v", *pricing.Properties.PricingTier)
		}
		if pricing.Properties.SubPlan == nil || *pricing.Properties.SubPlan != "DefenderForStorageV2" {
			t.Errorf("expected SubPlan DefenderForStorageV2, got %v", pricing.Properties.SubPlan)
		}

		extMap := make(map[string]*armsecurity.Extension, len(pricing.Properties.Extensions))
		for _, e := range pricing.Properties.Extensions {
			extMap[*e.Name] = e
		}

		malware, ok := extMap["OnUploadMalwareScanning"]
		if !ok {
			t.Fatal("expected OnUploadMalwareScanning extension")
		}
		if *malware.IsEnabled != armsecurity.IsEnabledTrue {
			t.Errorf("expected OnUploadMalwareScanning enabled")
		}
		if malware.AdditionalExtensionProperties["CapGBPerMonthPerStorageAccount"] != "10000" {
			t.Errorf("expected CapGBPerMonthPerStorageAccount=10000, got %v", malware.AdditionalExtensionProperties)
		}

		sensitive, ok := extMap["SensitiveDataDiscovery"]
		if !ok {
			t.Fatal("expected SensitiveDataDiscovery extension")
		}
		if *sensitive.IsEnabled != armsecurity.IsEnabledTrue {
			t.Errorf("expected SensitiveDataDiscovery enabled")
		}

		disabled, ok := extMap["SomeDisabledExt"]
		if !ok {
			t.Fatal("expected SomeDisabledExt extension")
		}
		if *disabled.IsEnabled != armsecurity.IsEnabledFalse {
			t.Errorf("expected SomeDisabledExt disabled")
		}
	})
}

func extWithStatus(name string, code string, message string) *armsecurity.Extension {
	msg := message
	return &armsecurity.Extension{
		Name: &name,
		OperationStatus: &armsecurity.OperationStatus{
			Code:    &code,
			Message: &msg,
		},
	}
}

func TestPlanSpec_ExtensionErrors(t *testing.T) {
	plan := PlanSpec{PlanName: "CloudPosture"}

	tests := []struct {
		name        string
		resp        armsecurity.PricingsClientUpdateResponse
		wantErr     bool
		wantContain []string
	}{
		{
			name:    "nil properties returns no error",
			resp:    armsecurity.PricingsClientUpdateResponse{},
			wantErr: false,
		},
		{
			name: "no extensions returns no error",
			resp: armsecurity.PricingsClientUpdateResponse{
				Pricing: armsecurity.Pricing{
					Properties: &armsecurity.PricingProperties{},
				},
			},
			wantErr: false,
		},
		{
			name: "all extensions succeeded returns no error",
			resp: armsecurity.PricingsClientUpdateResponse{
				Pricing: armsecurity.Pricing{
					Properties: &armsecurity.PricingProperties{
						Extensions: []*armsecurity.Extension{
							extWithStatus("AgentlessVmScanning", "Succeeded", ""),
							extWithStatus("SensitiveDataDiscovery", "Succeeded", ""),
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "one extension failed returns error with name and message",
			resp: armsecurity.PricingsClientUpdateResponse{
				Pricing: armsecurity.Pricing{
					Properties: &armsecurity.PricingProperties{
						Extensions: []*armsecurity.Extension{
							extWithStatus("AgentlessVmScanning", "Failed", "Failed assigning roles"),
						},
					},
				},
			},
			wantErr:     true,
			wantContain: []string{"AgentlessVmScanning", "Failed assigning roles"},
		},
		{
			name: "multiple extensions failed returns all in error",
			resp: armsecurity.PricingsClientUpdateResponse{
				Pricing: armsecurity.Pricing{
					Properties: &armsecurity.PricingProperties{
						Extensions: []*armsecurity.Extension{
							extWithStatus("AgentlessVmScanning", "Failed", "role error"),
							extWithStatus("SensitiveDataDiscovery", "Failed", "another error"),
						},
					},
				},
			},
			wantErr:     true,
			wantContain: []string{"AgentlessVmScanning", "SensitiveDataDiscovery"},
		},
		{
			name: "extension with nil operationStatus is skipped",
			resp: armsecurity.PricingsClientUpdateResponse{
				Pricing: armsecurity.Pricing{
					Properties: &armsecurity.PricingProperties{
						Extensions: []*armsecurity.Extension{
							{Name: func() *string { s := "AgentlessVmScanning"; return &s }()},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plan.extensionErrors(tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionErrors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				for _, sub := range tt.wantContain {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("extensionErrors() error %q does not contain %q", err.Error(), sub)
					}
				}
			}
		})
	}
}
