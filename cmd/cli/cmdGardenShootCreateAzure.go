package main

import (
	"context"
	"fmt"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	azuregardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/gardener"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

var (
	azureClientId       string
	azureClientSecret   string
	azureSubscriptionId string
	azureTenantID       string
	azureRegion         string
)

func init() {
	cmdGardenShootCreateAzure.Flags().StringVarP(&azureClientId, "clientId", "u", "", "Azure Client ID")
	cmdGardenShootCreateAzure.Flags().StringVarP(&azureClientSecret, "clientSecret", "p", "", "Azure Client Secret")
	cmdGardenShootCreateAzure.Flags().StringVarP(&azureSubscriptionId, "subscription", "", "westeurope", "Azure Subscription ID")
	cmdGardenShootCreateAzure.Flags().StringVarP(&azureTenantID, "tenant", "", "", "Azure Tenant ID")
	cmdGardenShootCreateAzure.Flags().StringVarP(&azureRegion, "region", "r", "eu-west-1", "Shoot region")

	_ = cmdGardenShootCreateAzure.MarkFlagRequired("clientId")
	_ = cmdGardenShootCreateAzure.MarkFlagRequired("clientSecret")
	_ = cmdGardenShootCreateAzure.MarkFlagRequired("subscription")
	_ = cmdGardenShootCreateAzure.MarkFlagRequired("tenant")
	cmdGardenShootCreate.AddCommand(cmdGardenShootCreateAzure)
}

var cmdGardenShootCreateAzure = &cobra.Command{
	Use: "azure",
	RunE: func(cmd *cobra.Command, args []string) error {
		mustAll(
			requiredShootName(),
			defaultGardenNamespace(),
		)

		c := helper.NewGardenClient()

		shoot := &gardenertypes.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
			Spec: gardenertypes.ShootSpec{
				Region:           azureRegion,
				CloudProfileName: ptr.To("az"),
				Networking: &gardenertypes.Networking{
					IPFamilies: []gardenertypes.IPFamily{gardenertypes.IPFamilyIPv4},
					Nodes:      ptr.To("10.250.0.0/22"),
					Pods:       ptr.To("100.64.0.0/12"),
					Services:   ptr.To("100.104.0.0/13"),
				},
				Provider: gardenertypes.Provider{
					Type: "azure",
					InfrastructureConfig: &runtime.RawExtension{
						Object: &azuregardener.InfrastructureConfig{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InfrastructureConfig",
								APIVersion: "azure.provider.extensions.gardener.cloud/v1alpha1",
							},
							Networks: azuregardener.NetworkConfig{
								VNet: azuregardener.VNet{
									CIDR: ptr.To("10.180.0.0/16"),
								},
								Zones: []azuregardener.Zone{
									{
										Name: 1,
										CIDR: "10.250.0.0/25",
									},
									{
										Name: 2,
										CIDR: "10.250.0.128/25",
									},
									{
										Name: 3,
										CIDR: "10.250.1.0/25",
									},
								},
							},
						},
					},
				},
				SecretBindingName: ptr.To(shootName),
			},
		}

		err := c.Create(context.Background(), shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
		fmt.Println("Created Shoot")

		sb := &gardenertypes.SecretBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
			Provider: &gardenertypes.SecretBindingProvider{
				Type: "azure",
			},
			SecretRef: corev1.SecretReference{
				Name:      shootName,
				Namespace: namespace,
			},
		}
		err = c.Create(context.Background(), sb)
		if err != nil {
			return fmt.Errorf("error creating SecretBinding: %w", err)
		}
		fmt.Println("Created SecretBinding")

		s := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
			StringData: map[string]string{
				"clientID":       azureClientId,
				"clientSecret":   azureClientSecret,
				"subscriptionID": azureSubscriptionId,
				"tenantID":       azureTenantID,
			},
		}
		err = c.Create(context.Background(), s)
		if err != nil {
			return fmt.Errorf("error creating Secret: %w", err)
		}
		fmt.Println("Created Secret")

		return nil
	},
}
