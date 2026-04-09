package v1beta1

func NewAwsWebAclBuilder() *AwsWebAclBuilder {
	return &AwsWebAclBuilder{
		AwsWebAcl: AwsWebAcl{
			Spec: AwsWebAclSpec{},
		},
	}
}

type AwsWebAclBuilder struct {
	AwsWebAcl AwsWebAcl
}

func (b *AwsWebAclBuilder) Reset() *AwsWebAclBuilder {
	b.AwsWebAcl = AwsWebAcl{}
	return b
}

func (b *AwsWebAclBuilder) WithDefaultAction(action AwsWebAclDefaultAction) *AwsWebAclBuilder {
	b.AwsWebAcl.Spec.DefaultAction = action
	return b
}

func (b *AwsWebAclBuilder) WithDescription(description string) *AwsWebAclBuilder {
	b.AwsWebAcl.Spec.Description = description
	return b
}

func (b *AwsWebAclBuilder) WithVisibilityConfig(config *AwsWebAclVisibilityConfig) *AwsWebAclBuilder {
	b.AwsWebAcl.Spec.VisibilityConfig = config
	return b
}

func (b *AwsWebAclBuilder) WithRule(rule AwsWebAclRule) *AwsWebAclBuilder {
	b.AwsWebAcl.Spec.Rules = append(b.AwsWebAcl.Spec.Rules, rule)
	return b
}

func (b *AwsWebAclBuilder) WithRules(rules []AwsWebAclRule) *AwsWebAclBuilder {
	b.AwsWebAcl.Spec.Rules = rules
	return b
}

// DefaultActionAllow returns a simple Allow default action
func DefaultActionAllow() AwsWebAclDefaultAction {
	return AwsWebAclDefaultAction{
		Allow: &AwsWebAclAllowAction{},
	}
}

// DefaultActionBlock returns a simple Block default action
func DefaultActionBlock() AwsWebAclDefaultAction {
	return AwsWebAclDefaultAction{
		Block: &AwsWebAclBlockAction{},
	}
}

// RuleActionAllow returns a simple Allow rule action
func RuleActionAllow() *AwsWebAclRuleActionType {
	return &AwsWebAclRuleActionType{
		Allow: &AwsWebAclAllowAction{},
	}
}

// RuleActionBlock returns a simple Block rule action
func RuleActionBlock() *AwsWebAclRuleActionType {
	return &AwsWebAclRuleActionType{
		Block: &AwsWebAclBlockAction{},
	}
}

// RuleActionCount returns a simple Count rule action
func RuleActionCount() *AwsWebAclRuleActionType {
	return &AwsWebAclRuleActionType{
		Count: &AwsWebAclCountAction{},
	}
}

// RuleActionCaptcha returns a simple Captcha rule action
func RuleActionCaptcha() *AwsWebAclRuleActionType {
	return &AwsWebAclRuleActionType{
		Captcha: &AwsWebAclCaptchaAction{},
	}
}

// OverrideActionNone returns a None override action (don't override)
func OverrideActionNone() *AwsWebAclOverrideAction {
	return &AwsWebAclOverrideAction{
		None: &AwsWebAclNoneAction{},
	}
}

// OverrideActionCount returns a Count override action (override all to count)
func OverrideActionCount() *AwsWebAclOverrideAction {
	return &AwsWebAclOverrideAction{
		Count: &AwsWebAclCountAction{},
	}
}
