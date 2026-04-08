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
