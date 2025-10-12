package fake

type ClusterBuilder struct {
	cluster *Cluster
}

func NewClusterBuilder() *ClusterBuilder {
	return (&ClusterBuilder{}).Reset()
}

func (b *ClusterBuilder) Reset() *ClusterBuilder {
	b.cluster = &Cluster{}
	b.cluster.Cache = &Cache{}

	return b
}

func (b *ClusterBuilder) WithStartErrors(pre error, post error) *ClusterBuilder {
	b.cluster.ErrPre = pre
	b.cluster.ErrPost = post
	return b
}

func (b *ClusterBuilder) WithCache(started bool, synced bool) *ClusterBuilder {
	b.cluster.Cache.Started = started
	b.cluster.Cache.Synced = synced
	return b
}

func (b *ClusterBuilder) Build() *Cluster {
	return b.cluster
}
