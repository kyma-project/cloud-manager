package awswebacl

import (
	"testing"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/stretchr/testify/assert"
)

func TestCompareDefaultAction(t *testing.T) {
	t.Run("Both nil", func(t *testing.T) {
		result := compareDefaultAction(nil, nil)
		assert.True(t, result)
	})

	t.Run("One nil", func(t *testing.T) {
		result := compareDefaultAction(&wafv2types.DefaultAction{
			Allow: &wafv2types.AllowAction{},
		}, nil)
		assert.False(t, result)
	})

	t.Run("Both Allow", func(t *testing.T) {
		aws := &wafv2types.DefaultAction{Allow: &wafv2types.AllowAction{}}
		spec := &wafv2types.DefaultAction{Allow: &wafv2types.AllowAction{}}
		result := compareDefaultAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Both Block", func(t *testing.T) {
		aws := &wafv2types.DefaultAction{Block: &wafv2types.BlockAction{}}
		spec := &wafv2types.DefaultAction{Block: &wafv2types.BlockAction{}}
		result := compareDefaultAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Different actions - Allow vs Block", func(t *testing.T) {
		aws := &wafv2types.DefaultAction{Allow: &wafv2types.AllowAction{}}
		spec := &wafv2types.DefaultAction{Block: &wafv2types.BlockAction{}}
		result := compareDefaultAction(aws, spec)
		assert.False(t, result)
	})
}

func TestCompareVisibilityConfig(t *testing.T) {
	t.Run("Both nil", func(t *testing.T) {
		result := compareVisibilityConfig(nil, nil)
		assert.True(t, result)
	})

	t.Run("One nil", func(t *testing.T) {
		result := compareVisibilityConfig(&wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test"),
			SampledRequestsEnabled:   false,
		}, nil)
		assert.False(t, result)
	})

	t.Run("Identical configs", func(t *testing.T) {
		aws := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test-metric"),
			SampledRequestsEnabled:   false,
		}
		spec := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test-metric"),
			SampledRequestsEnabled:   false,
		}
		result := compareVisibilityConfig(aws, spec)
		assert.True(t, result)
	})

	t.Run("Different CloudWatchMetricsEnabled", func(t *testing.T) {
		aws := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test"),
			SampledRequestsEnabled:   false,
		}
		spec := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: false,
			MetricName:               new("test"),
			SampledRequestsEnabled:   false,
		}
		result := compareVisibilityConfig(aws, spec)
		assert.False(t, result)
	})

	t.Run("Different SampledRequestsEnabled", func(t *testing.T) {
		aws := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test"),
			SampledRequestsEnabled:   true,
		}
		spec := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test"),
			SampledRequestsEnabled:   false,
		}
		result := compareVisibilityConfig(aws, spec)
		assert.False(t, result)
	})

	t.Run("Different MetricName", func(t *testing.T) {
		aws := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("metric-1"),
			SampledRequestsEnabled:   false,
		}
		spec := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("metric-2"),
			SampledRequestsEnabled:   false,
		}
		result := compareVisibilityConfig(aws, spec)
		assert.False(t, result)
	})

	t.Run("MetricName nil vs non-nil", func(t *testing.T) {
		aws := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               nil,
			SampledRequestsEnabled:   false,
		}
		spec := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("metric"),
			SampledRequestsEnabled:   false,
		}
		result := compareVisibilityConfig(aws, spec)
		assert.False(t, result)
	})
}

func TestCompareRuleAction(t *testing.T) {
	t.Run("Both nil", func(t *testing.T) {
		result := compareRuleAction(nil, nil)
		assert.True(t, result)
	})

	t.Run("One nil", func(t *testing.T) {
		result := compareRuleAction(&wafv2types.RuleAction{
			Allow: &wafv2types.AllowAction{},
		}, nil)
		assert.False(t, result)
	})

	t.Run("Both Allow", func(t *testing.T) {
		aws := &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}}
		spec := &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}}
		result := compareRuleAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Both Block", func(t *testing.T) {
		aws := &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}}
		spec := &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}}
		result := compareRuleAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Both Count", func(t *testing.T) {
		aws := &wafv2types.RuleAction{Count: &wafv2types.CountAction{}}
		spec := &wafv2types.RuleAction{Count: &wafv2types.CountAction{}}
		result := compareRuleAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Both Captcha", func(t *testing.T) {
		aws := &wafv2types.RuleAction{Captcha: &wafv2types.CaptchaAction{}}
		spec := &wafv2types.RuleAction{Captcha: &wafv2types.CaptchaAction{}}
		result := compareRuleAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Different actions - Allow vs Block", func(t *testing.T) {
		aws := &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}}
		spec := &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}}
		result := compareRuleAction(aws, spec)
		assert.False(t, result)
	})

	t.Run("Different actions - Count vs Captcha", func(t *testing.T) {
		aws := &wafv2types.RuleAction{Count: &wafv2types.CountAction{}}
		spec := &wafv2types.RuleAction{Captcha: &wafv2types.CaptchaAction{}}
		result := compareRuleAction(aws, spec)
		assert.False(t, result)
	})
}

func TestCompareOverrideAction(t *testing.T) {
	t.Run("Both nil", func(t *testing.T) {
		result := compareOverrideAction(nil, nil)
		assert.True(t, result)
	})

	t.Run("One nil", func(t *testing.T) {
		result := compareOverrideAction(&wafv2types.OverrideAction{
			None: &wafv2types.NoneAction{},
		}, nil)
		assert.False(t, result)
	})

	t.Run("Both None", func(t *testing.T) {
		aws := &wafv2types.OverrideAction{None: &wafv2types.NoneAction{}}
		spec := &wafv2types.OverrideAction{None: &wafv2types.NoneAction{}}
		result := compareOverrideAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Both Count", func(t *testing.T) {
		aws := &wafv2types.OverrideAction{Count: &wafv2types.CountAction{}}
		spec := &wafv2types.OverrideAction{Count: &wafv2types.CountAction{}}
		result := compareOverrideAction(aws, spec)
		assert.True(t, result)
	})

	t.Run("Different - None vs Count", func(t *testing.T) {
		aws := &wafv2types.OverrideAction{None: &wafv2types.NoneAction{}}
		spec := &wafv2types.OverrideAction{Count: &wafv2types.CountAction{}}
		result := compareOverrideAction(aws, spec)
		assert.False(t, result)
	})
}

func TestCompareRule(t *testing.T) {
	t.Run("Both nil", func(t *testing.T) {
		result := compareRule(nil, nil)
		assert.True(t, result)
	})

	t.Run("One nil", func(t *testing.T) {
		result := compareRule(&wafv2types.Rule{
			Name:     new("test"),
			Priority: 10,
		}, nil)
		assert.False(t, result)
	})

	t.Run("Identical rules with action", func(t *testing.T) {
		statement := &wafv2types.Statement{
			GeoMatchStatement: &wafv2types.GeoMatchStatement{
				CountryCodes: []wafv2types.CountryCode{wafv2types.CountryCodeUs},
			},
		}
		visConfig := &wafv2types.VisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               new("test"),
			SampledRequestsEnabled:   false,
		}

		aws := &wafv2types.Rule{
			Name:             new("test-rule"),
			Priority:         10,
			Action:           &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}},
			Statement:        statement,
			VisibilityConfig: visConfig,
		}
		spec := &wafv2types.Rule{
			Name:             new("test-rule"),
			Priority:         10,
			Action:           &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}},
			Statement:        statement,
			VisibilityConfig: visConfig,
		}
		result := compareRule(aws, spec)
		assert.True(t, result)
	})

	t.Run("Different names", func(t *testing.T) {
		aws := &wafv2types.Rule{
			Name:     new("rule-1"),
			Priority: 10,
		}
		spec := &wafv2types.Rule{
			Name:     new("rule-2"),
			Priority: 10,
		}
		result := compareRule(aws, spec)
		assert.False(t, result)
	})

	t.Run("Different priorities", func(t *testing.T) {
		aws := &wafv2types.Rule{
			Name:     new("test-rule"),
			Priority: 10,
		}
		spec := &wafv2types.Rule{
			Name:     new("test-rule"),
			Priority: 20,
		}
		result := compareRule(aws, spec)
		assert.False(t, result)
	})

	t.Run("Different actions", func(t *testing.T) {
		aws := &wafv2types.Rule{
			Name:     new("test-rule"),
			Priority: 10,
			Action:   &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}},
		}
		spec := &wafv2types.Rule{
			Name:     new("test-rule"),
			Priority: 10,
			Action:   &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}},
		}
		result := compareRule(aws, spec)
		assert.False(t, result)
	})

	t.Run("Identical rules with override action (managed rule group)", func(t *testing.T) {
		statement := &wafv2types.Statement{
			ManagedRuleGroupStatement: &wafv2types.ManagedRuleGroupStatement{
				VendorName: new("AWS"),
				Name:       new("AWSManagedRulesCommonRuleSet"),
			},
		}

		aws := &wafv2types.Rule{
			Name:           new("managed-rule"),
			Priority:       5,
			OverrideAction: &wafv2types.OverrideAction{None: &wafv2types.NoneAction{}},
			Statement:      statement,
		}
		spec := &wafv2types.Rule{
			Name:           new("managed-rule"),
			Priority:       5,
			OverrideAction: &wafv2types.OverrideAction{None: &wafv2types.NoneAction{}},
			Statement:      statement,
		}
		result := compareRule(aws, spec)
		assert.True(t, result)
	})
}

func TestCompareRules(t *testing.T) {
	t.Run("Both empty", func(t *testing.T) {
		result := compareRules([]wafv2types.Rule{}, []wafv2types.Rule{})
		assert.True(t, result)
	})

	t.Run("Different lengths", func(t *testing.T) {
		aws := []wafv2types.Rule{
			{Name: new("rule-1"), Priority: 1},
		}
		spec := []wafv2types.Rule{
			{Name: new("rule-1"), Priority: 1},
			{Name: new("rule-2"), Priority: 2},
		}
		result := compareRules(aws, spec)
		assert.False(t, result)
	})

	t.Run("Same length, identical rules", func(t *testing.T) {
		statement := &wafv2types.Statement{
			GeoMatchStatement: &wafv2types.GeoMatchStatement{
				CountryCodes: []wafv2types.CountryCode{wafv2types.CountryCodeUs},
			},
		}

		aws := []wafv2types.Rule{
			{
				Name:      new("rule-1"),
				Priority:  1,
				Action:    &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}},
				Statement: statement,
			},
			{
				Name:      new("rule-2"),
				Priority:  2,
				Action:    &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}},
				Statement: statement,
			},
		}
		spec := []wafv2types.Rule{
			{
				Name:      new("rule-1"),
				Priority:  1,
				Action:    &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}},
				Statement: statement,
			},
			{
				Name:      new("rule-2"),
				Priority:  2,
				Action:    &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}},
				Statement: statement,
			},
		}
		result := compareRules(aws, spec)
		assert.True(t, result)
	})

	t.Run("Same length, different rules", func(t *testing.T) {
		statement := &wafv2types.Statement{
			GeoMatchStatement: &wafv2types.GeoMatchStatement{
				CountryCodes: []wafv2types.CountryCode{wafv2types.CountryCodeUs},
			},
		}

		aws := []wafv2types.Rule{
			{
				Name:      new("rule-1"),
				Priority:  1,
				Action:    &wafv2types.RuleAction{Allow: &wafv2types.AllowAction{}},
				Statement: statement,
			},
		}
		spec := []wafv2types.Rule{
			{
				Name:      new("rule-1"),
				Priority:  1,
				Action:    &wafv2types.RuleAction{Block: &wafv2types.BlockAction{}}, // Different action
				Statement: statement,
			},
		}
		result := compareRules(aws, spec)
		assert.False(t, result)
	})
}
