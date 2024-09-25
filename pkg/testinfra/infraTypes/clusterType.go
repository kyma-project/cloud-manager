package infraTypes

type ClusterType string

const (
	ClusterTypeKcp    = ClusterType("kcp")
	ClusterTypeSkr    = ClusterType("skr")
	ClusterTypeGarden = ClusterType("garden")
)
