package security

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
)

type ExtensionSpec struct {
	Name                          string
	AdditionalExtensionProperties map[string]any
}

type PlanSpec struct {
	PlanName           string
	SubPlan            string
	RequiredExtensions []ExtensionSpec
	DisabledExtensions []string
}

func (p PlanSpec) NeedsUpdate(current *armsecurity.Pricing, targetTier armsecurity.PricingTier) bool {
	if current == nil || current.Properties == nil || current.Properties.PricingTier == nil {
		return true
	}
	if *current.Properties.PricingTier != targetTier {
		return true
	}
	if targetTier == armsecurity.PricingTierFree {
		return false
	}
	if p.SubPlan != "" {
		if current.Properties.SubPlan == nil || *current.Properties.SubPlan != p.SubPlan {
			return true
		}
	}
	return !p.hasAllExtensionsEnabled(current.Properties.Extensions)
}

func (p PlanSpec) BuildPricing(targetTier armsecurity.PricingTier) armsecurity.Pricing {
	pricing := armsecurity.Pricing{
		Properties: &armsecurity.PricingProperties{
			PricingTier: &targetTier,
		},
	}
	if targetTier == armsecurity.PricingTierStandard {
		if p.SubPlan != "" {
			subPlan := p.SubPlan
			pricing.Properties.SubPlan = &subPlan
		}
		pricing.Properties.Extensions = p.buildExtensions()
	}
	return pricing
}

func (p PlanSpec) extensionErrors(resp armsecurity.PricingsClientUpdateResponse) error {
	if resp.Properties == nil {
		return nil
	}
	var msgs []string
	for _, ext := range resp.Properties.Extensions {
		if ext.Name == nil || ext.OperationStatus == nil || ext.OperationStatus.Code == nil {
			continue
		}
		if *ext.OperationStatus.Code == "Failed" {
			msg := *ext.Name
			if ext.OperationStatus.Message != nil {
				msg += ": " + *ext.OperationStatus.Message
			}
			msgs = append(msgs, msg)
		}
	}
	if len(msgs) > 0 {
		return fmt.Errorf("extension operation failures: %s", strings.Join(msgs, "; "))
	}
	return nil
}

func (p PlanSpec) hasAllExtensionsEnabled(current []*armsecurity.Extension) bool {
	enabled := make(map[string]bool, len(current))
	for _, ext := range current {
		if ext.Name != nil && ext.IsEnabled != nil && *ext.IsEnabled == armsecurity.IsEnabledTrue {
			enabled[*ext.Name] = true
		}
	}
	for _, spec := range p.RequiredExtensions {
		if !enabled[spec.Name] {
			return false
		}
	}
	return true
}

func (p PlanSpec) buildExtensions() []*armsecurity.Extension {
	result := make([]*armsecurity.Extension, 0, len(p.RequiredExtensions)+len(p.DisabledExtensions))
	isTrue := armsecurity.IsEnabledTrue
	isFalse := armsecurity.IsEnabledFalse
	for _, spec := range p.RequiredExtensions {
		n := spec.Name
		ext := &armsecurity.Extension{Name: &n, IsEnabled: &isTrue}
		if spec.AdditionalExtensionProperties != nil {
			ext.AdditionalExtensionProperties = spec.AdditionalExtensionProperties
		}
		result = append(result, ext)
	}
	for _, name := range p.DisabledExtensions {
		n := name
		result = append(result, &armsecurity.Extension{Name: &n, IsEnabled: &isFalse})
	}
	return result
}
