package dsl

import "sigs.k8s.io/controller-runtime/pkg/client"

type ObjAssertion func(obj client.Object) error

type ObjAssertions []ObjAssertion

func NewObjAssertions(items []ObjAssertion) ObjAssertions {
	return append(ObjAssertions{}, items...)
}

func (x ObjAssertions) AssertObj(obj client.Object) error {
	for _, a := range x {
		if err := a(obj); err != nil {
			return err
		}
	}
	return nil
}
