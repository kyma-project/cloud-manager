package util

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Apply(ctx context.Context, clnt client.Client, objects []*unstructured.Unstructured) error {
	for _, desiredObj := range objects {
		key := client.ObjectKeyFromObject(desiredObj)
		existingObj := desiredObj.DeepCopyObject().(*unstructured.Unstructured)
		err := clnt.Get(ctx, key, existingObj)
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("error getting existing crd %s: %w", key, err)
		}
		if err != nil {
			// does not exist, should be created
			err = clnt.Create(ctx, desiredObj)
			if err != nil {
				return fmt.Errorf("error creating %s %s: %w", desiredObj.GetKind(), key, err)
			}
		} else {
			// exists, should be updated
			// get metadata from existing
			metadata, _, err := unstructured.NestedMap(existingObj.Object, "metadata")
			if err != nil {
				return fmt.Errorf("error getting existing %s metadata: %w", key, err)
			}
			// set existing metadata to desired
			err = unstructured.SetNestedField(desiredObj.Object, metadata, "metadata")
			if err != nil {
				return fmt.Errorf("error setting desired %s metadata: %w", key, err)
			}
			// update desired
			err = clnt.Update(ctx, desiredObj)
			if err != nil {
				return fmt.Errorf("error updating existing %s: %w", key, err)
			}
		}
	}
	return nil
}
