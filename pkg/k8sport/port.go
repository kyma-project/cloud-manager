package k8sport

import composedv2 "github.com/kyma-project/cloud-manager/pkg/composed/v2"

type K8sPort interface {
	K8sEventPort
	K8sLoadPort
	K8sLabelObjPort
	K8sAnnotateObjPort
	K8sCreatePort
	K8sDeletePort

	ClusterId() string
}

type k8sPort struct {
	K8sEventPort
	K8sLoadPort
	K8sLabelObjPort
	K8sAnnotateObjPort
	K8sCreatePort
	K8sDeletePort

	clusterId string
}

func (p *k8sPort) ClusterId() string {
	return p.clusterId
}

func NewK8sPort(clusterID string) K8sPort {
	return &k8sPort{
		clusterId:          clusterID,
		K8sEventPort:       NewK8sEventPort(clusterID),
		K8sLoadPort:        NewK8sLoadPort(clusterID),
		K8sLabelObjPort:    NewK8sLabelObjPort(clusterID),
		K8sAnnotateObjPort: NewK8sAnnotateObjPort(clusterID),
		K8sCreatePort:      NewK8sCreatePort(clusterID),
		K8sDeletePort:      NewK8sDeletePort(clusterID),
	}
}

func NewK8sPortOnDefaultCluster() K8sPort {
	return NewK8sPort(composedv2.DefaultClusterID)
}
