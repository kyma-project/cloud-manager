package testinfra

import (
	"fmt"
	"strings"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (i *infra) FinalizerReport() {
	groupsToCheck := map[string]struct{}{
		cloudcontrolv1beta1.GroupVersion.Group:     {},
		cloudresourcesv1beta1.GroupVersion.Group:   {},
		gardenerapicore.GroupName:                  {},
		infrastructuremanagerv1.GroupVersion.Group: {},
		operatorv1beta2.GroupVersion.Group:         {},
	}
	kindsToCheck := map[string]struct{}{
		"ConfigMap":             {},
		"Secret":                {},
		"PersistentVolume":      {},
		"PersistentVolumeClaim": {},
	}
	fmt.Println("Finalizer Report Start =======================================")
	for name, clsrt := range i.clusters {
		fmt.Printf("Cluster %s\n", name)
		for gvk := range clsrt.Scheme().AllKnownTypes() {
			_, groupOK := groupsToCheck[gvk.Group]
			_, kindOK := kindsToCheck[gvk.Kind]
			if !groupOK && !kindOK {
				continue
			}
			if strings.HasSuffix(gvk.Kind, "List") {
				continue
			}
			if strings.HasSuffix(gvk.Kind, "Options") {
				continue
			}

			fmt.Printf("  %s\n", gvk.String())

			list := &unstructured.UnstructuredList{}
			list.SetGroupVersionKind(gvk.GroupVersion().WithKind(gvk.Kind + "List"))
			err := clsrt.Client().List(i.Ctx(), list)
			err = util.IgnoreNoMatch(err)
			if err != nil {
				fmt.Printf("%T\n", err)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf("error listing GVK %s in cluster %s", gvk.String(), name))
			}

			var arrWithFinalizer []string
			for _, obj := range list.Items {
				finalizers := obj.GetFinalizers()
				if len(finalizers) > 0 {
					arrWithFinalizer = append(arrWithFinalizer, client.ObjectKeyFromObject(&obj).String())
				}
			}

			if len(arrWithFinalizer) > 0 {
				for _, item := range arrWithFinalizer {
					fmt.Printf("   * %s\n", item)
				}
				fmt.Println("   FINALIZERS PRESENT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			}
		}
	}
	fmt.Println("Finalizer Report End =======================================")
}
