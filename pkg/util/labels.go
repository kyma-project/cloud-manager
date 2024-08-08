package util

import "regexp"

const (
	WellKnownK8sLabelName      = "app.kubernetes.io/name"
	WellKnownK8sLabelInstance  = "app.kubernetes.io/instance"
	WellKnownK8sLabelVersion   = "app.kubernetes.io/version"
	WellKnownK8sLabelComponent = "app.kubernetes.io/component"
	WellKnownK8sLabelPartOf    = "app.kubernetes.io/part-of"
	WellKnownK8sLabelManagedBy = "app.kubernetes.io/managed-by"
)

const (
	DefaultCloudManagerComponentLabelValue = "cloud-manager"
	DefaultCloudManagerPartOfLabelValue    = "kyma"
	DefaultCloudManagerManagedByLabelValue = "cloud-manager"
)

// Keys and values can contain only lowercase letters, numeric characters, underscores, and dashes.
// All characters must use UTF-8 encoding, and international characters are allowed. Keys must start with a lowercase letter or international character.
// https://cloud.google.com/bigtable/docs/creating-managing-labels
const (
	GcpLabelKcpName   = "kyma-project__cloud-manager__kcp-name"
	GcpLabelSkrName   = "kyma-project__cloud-manager__skr-name"
	GcpLabelScopeName = "kyma-project__cloud-manager__scope-name"
	GcpLabelShootName = "kyma-project__cloud-manager__shoot-name"
)

var _ LabelBuilder = &labelBuilder{}

type LabelBuilder interface {
	WithName(name string) LabelBuilder
	WithInstance(instance string) LabelBuilder
	WithVersion(version string) LabelBuilder
	WithComponent(component string) LabelBuilder
	WithPartOf(partOf string) LabelBuilder
	WithManagedBy(managedBy string) LabelBuilder
	WithCustomLabel(labelName, labelValue string) LabelBuilder
	WithCustomLabels(customLabels map[string]string) LabelBuilder
	WithCloudManagerDefaults() LabelBuilder
	WithGcpLabels(scopeName, shootName string) LabelBuilder

	// Returns map[string]string that reflects the deep copy of provided building blocks
	Build() map[string]string
}

type labelBuilder struct {
	labels map[string]string
}

func (labelBuilder *labelBuilder) WithName(name string) LabelBuilder {
	labelBuilder.labels[WellKnownK8sLabelName] = name
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithInstance(instance string) LabelBuilder {
	labelBuilder.labels[WellKnownK8sLabelInstance] = instance
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithVersion(version string) LabelBuilder {
	labelBuilder.labels[WellKnownK8sLabelVersion] = version
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithComponent(component string) LabelBuilder {
	labelBuilder.labels[WellKnownK8sLabelComponent] = component
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithPartOf(partOf string) LabelBuilder {
	labelBuilder.labels[WellKnownK8sLabelPartOf] = partOf
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithManagedBy(managedBy string) LabelBuilder {
	labelBuilder.labels[WellKnownK8sLabelManagedBy] = managedBy
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithCustomLabel(labelName, labelValue string) LabelBuilder {
	labelBuilder.labels[labelName] = labelValue
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithCustomLabels(customLabels map[string]string) LabelBuilder {
	for labelName, labelValue := range customLabels {
		labelBuilder.WithCustomLabel(labelName, labelValue)
	}
	return labelBuilder
}

func (labelBuilder *labelBuilder) WithCloudManagerDefaults() LabelBuilder {
	return labelBuilder.WithComponent(DefaultCloudManagerComponentLabelValue).WithPartOf(DefaultCloudManagerPartOfLabelValue).WithManagedBy(DefaultCloudManagerManagedByLabelValue)
}

func (labelBuilder *labelBuilder) WithGcpLabels(scopeName, shootName string) LabelBuilder {
	re := regexp.MustCompile(`^[a-z\p{L}][a-z0-9\p{L}_-]*$`)

	if re.MatchString(scopeName) {
		labelBuilder.labels[GcpLabelScopeName] = scopeName
	}

	if re.MatchString(shootName) {
		labelBuilder.labels[GcpLabelShootName] = shootName
	}

	return labelBuilder
}

func (labelBuilder *labelBuilder) Build() map[string]string {
	resultMap := make(map[string]string)

	for key, value := range labelBuilder.labels {
		resultMap[key] = value
	}

	return resultMap
}

func NewLabelBuilder() LabelBuilder {
	return &labelBuilder{
		labels: make(map[string]string),
	}
}
