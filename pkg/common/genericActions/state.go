package genericActions

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources-manager/apis/cloud-resources/v1beta1"
	composed "github.com/kyma-project/cloud-resources-manager/pkg/common/composedAction"
)

func NewState(baseState composed.State) composed.State {
	return &state{
		State: baseState,
	}
}

type StateWithCloudResources interface {
	ServedCloudResources() *cloudresourcesv1beta1.CloudResources
	SetServedCloudResources(cr *cloudresourcesv1beta1.CloudResources)

	CloudResourcesList() *cloudresourcesv1beta1.CloudResourcesList
	SetCloudResourcesList(list *cloudresourcesv1beta1.CloudResourcesList)
}

type state struct {
	composed.State
	cloudResourcesList   *cloudresourcesv1beta1.CloudResourcesList
	servedCloudResources *cloudresourcesv1beta1.CloudResources
}

func (s *state) ServedCloudResources() *cloudresourcesv1beta1.CloudResources {
	return s.servedCloudResources
}

func (s *state) SetServedCloudResources(cr *cloudresourcesv1beta1.CloudResources) {
	s.servedCloudResources = cr
}

func (s *state) CloudResourcesList() *cloudresourcesv1beta1.CloudResourcesList {
	return s.cloudResourcesList
}

func (s *state) SetCloudResourcesList(list *cloudresourcesv1beta1.CloudResourcesList) {
	s.cloudResourcesList = list
}
