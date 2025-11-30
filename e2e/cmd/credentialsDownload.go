package main

import (
	"fmt"
	"os"
	"path/filepath"

	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var cmdCredentialsDownload = &cobra.Command{
	Use:     "download",
	Aliases: []string{"down", "d"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.CredentialsDir == "" {
			return fmt.Errorf("credentials dir is required")
		}
		if err := os.MkdirAll(config.CredentialsDir, 0755); err != nil {
			return fmt.Errorf("failed to create credentials dir: %w", err)
		}
		keb, err := e2ekeb.Create(rootCtx, config)
		if err != nil {
			return fmt.Errorf("failed to create keb: %w", err)
		}

		for secretName, mapping := range config.DownloadGardenSecrets {
			fmt.Printf("Downloading secret %s\n", secretName)
			secret := &corev1.Secret{}
			err = keb.GardenClient().Get(rootCtx, types.NamespacedName{
				Namespace: config.GardenNamespace,
				Name:      secretName,
			}, secret)
			if err != nil {
				return fmt.Errorf("failed to get secret %s: %w", secretName, err)
			}
			for filename, key := range mapping {
				fmt.Printf(" * %s\n", filename)
				val, ok := secret.Data[key]
				if !ok {
					fmt.Printf("key %s not found in secret %s\n", key, secretName)
					continue
				}
				err = os.WriteFile(filepath.Join(config.CredentialsDir, filename), val, 0644)
				if err != nil {
					return fmt.Errorf("failed to write file %s: %w", filename, err)
				}
			}
		}

		fmt.Println("Successfully downloaded all credentials")

		return nil
	},
}

func init() {
	cmdCredentials.AddCommand(cmdCredentialsDownload)
}
