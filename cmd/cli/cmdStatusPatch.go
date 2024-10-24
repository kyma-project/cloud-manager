package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	cmdStatusPatch.Flags().StringVarP(&patchOwner, "owner", "o", "cloud-manager", "The owner of the patch")
	cmdStatus.AddCommand(cmdStatusPatch)
}

var (
	patchOwner string
)

var cmdStatusPatch = &cobra.Command{
	Use:     "patch",
	Aliases: []string{"p"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateStatusFlags(); err != nil {
			return err
		}
		statusData, err := loadStatusFile()
		if err != nil {
			return err
		}

		clnt := helper.NewKcpClient()

		ctx := context.Background()

		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(schema.FromAPIVersionAndKind(apiVersion, kind))

		if err := clnt.Get(ctx, client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		}, obj); err != nil {
			return fmt.Errorf("could not get object: %w", err)
		}

		objToPatch := &unstructured.Unstructured{}
		objToPatch.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		objToPatch.SetName(name)
		objToPatch.SetNamespace(namespace)
		objToPatch.Object["status"] = statusData["status"]

		if err := clnt.Status().Patch(ctx, objToPatch, client.Apply, client.ForceOwnership, client.FieldOwner(patchOwner)); err != nil {
			return err
		}

		return nil
	},
}
