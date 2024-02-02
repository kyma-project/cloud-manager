package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/cmd/cli/helper"
	awsgardener "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/gardener"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

var (
	awsRegion  string
	awsProfile string
)

func init() {
	cmdGardenShootCreateAws.Flags().StringVarP(&awsRegion, "region", "r", "eu-west-1", "Shoot region")
	cmdGardenShootCreateAws.Flags().StringVarP(&awsProfile, "profile", "p", "", "AWS Access Key Secret")
	_ = cmdGardenShootCreateAws.MarkFlagRequired("profile")
	cmdGardenShootCreate.AddCommand(cmdGardenShootCreateAws)
}

var cmdGardenShootCreateAws = &cobra.Command{
	Use: "aws",
	RunE: func(cmd *cobra.Command, args []string) error {
		mustAll(
			requiredShootName(),
			defaultGardenNamespace(),
		)

		cfg, err := config.LoadDefaultConfig(
			context.Background(),
			config.WithSharedConfigProfile(awsProfile),
		)
		if err != nil {
			return fmt.Errorf("error loading aws profile %s: %s", awsProfile, err)
		}
		creds, err := cfg.Credentials.Retrieve(context.Background())
		if err != nil {
			return fmt.Errorf("error getting aws profile %s credentials: %w", awsProfile, err)
		}

		c := helper.NewGardenClient()

		shoot := &gardenerTypes.Shoot{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
			Spec: gardenerTypes.ShootSpec{
				Region: cfg.Region,
				Networking: &gardenerTypes.Networking{
					IPFamilies: []gardenerTypes.IPFamily{gardenerTypes.IPFamilyIPv4},
					Nodes:      pointer.String("10.180.0.0/16"),
					Pods:       pointer.String("100.64.0.0/12"),
					Services:   pointer.String("100.104.0.0/13"),
				},
				Provider: gardenerTypes.Provider{
					Type: "aws",
					InfrastructureConfig: &runtime.RawExtension{
						Object: &awsgardener.InfrastructureConfig{
							TypeMeta: metav1.TypeMeta{
								Kind:       "InfrastructureConfig",
								APIVersion: "aws.provider.extensions.gardener.cloud/v1alpha1",
							},
							Networks: awsgardener.Networks{
								VPC: awsgardener.VPC{
									CIDR: pointer.String("10.180.0.0/16"),
								},
								Zones: []awsgardener.Zone{
									{
										Name:     "eu-west-1a",
										Internal: "10.180.48.0/20",
										Public:   "10.180.32.0/20",
										Workers:  "10.180.0.0/19",
									},
									{
										Name:     "eu-west-1b",
										Internal: "10.180.112.0/20",
										Public:   "10.180.96.0/20",
										Workers:  "10.180.64.0/19",
									},
									{
										Name:     "eu-west-1c",
										Internal: "10.180.176.0/20",
										Public:   "10.180.160.0/20",
										Workers:  "10.180.128.0/19",
									},
								},
							},
						},
					},
				},
				SecretBindingName: pointer.String(shootName),
			},
		}
		err = c.Create(context.Background(), shoot)
		if err != nil {
			return fmt.Errorf("error creating Shoot: %w", err)
		}
		fmt.Println("Created Shoot")

		sb := &gardenerTypes.SecretBinding{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      shootName,
			},
			Provider: &gardenerTypes.SecretBindingProvider{
				Type: "aws",
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
				"accessKeyID":     creds.AccessKeyID,
				"secretAccessKey": creds.SecretAccessKey,
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
