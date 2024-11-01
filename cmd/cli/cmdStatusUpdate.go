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
	cmdStatus.AddCommand(cmdStatusUpdate)
}

var cmdStatusUpdate = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u"},
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

		for k, v := range statusData {
			obj.Object[k] = v
		}

		if err := clnt.Status().Update(ctx, obj); err != nil {
			return err
		}

		return nil
	},
}
