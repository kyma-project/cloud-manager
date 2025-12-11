package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/kyma-project/cloud-manager/e2e/cloud"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"github.com/spf13/cobra"
)

var (
	tfProviders     []string
	tfSource        string
	tfVariables     []string
	tfConfirmDelete bool
)

var cmdCloudTf = &cobra.Command{
	Use: "tf",
	RunE: func(cmd *cobra.Command, args []string) error {
		gardenClient, err := config.CreateGardenClient()
		if err != nil {
			return fmt.Errorf("failed to create garden client: %w", err)
		}

		cld, err := cloud.Create(rootCtx, gardenClient, config)
		if err != nil {
			return fmt.Errorf("failed to create cloud: %w", err)
		}

		// alias is not important here since used w/out cluster and resource declaration
		b := cld.WorkspaceBuilder(util.RandomString(8))
		for _, p := range tfProviders {
			b.WithProvider(p)
		}

		if strings.HasPrefix(tfSource, "./") || strings.HasPrefix(tfSource, "../") {
			tfSource = path.Join(config.ConfigDir, "e2e/tf", tfSource)
		}

		b.WithSource(tfSource)

		for _, v := range tfVariables {
			parts := strings.Split(v, "=")
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable format: %s", v)
			}
			b.WithVariable(parts[0], parts[1])
		}

		if err := b.Validate(); err != nil {
			return err
		}

		w := b.Build()

		fmt.Println("Destroying previous workspace, if any...")
		if err := w.Destroy(); err != nil {
			return fmt.Errorf("delete workspace: %w\n\n%s", err, w.Out())
		}

		fmt.Println("Create workspace...")
		if err := w.Create(); err != nil {
			return fmt.Errorf("create workspace: %w\n\n%s", err, w.Out())
		}

		fmt.Println("Init...")
		if err := w.Init(); err != nil {
			return fmt.Errorf("init workspace: %w\n\n%s", err, w.Out())
		}

		fmt.Println("Plan...")
		if err := w.Plan(); err != nil {
			return fmt.Errorf("plan workspace: %w\n\n%s", err, w.Out())
		}

		fmt.Println("Apply...")
		if err := w.Apply(); err != nil {
			return fmt.Errorf("apply workspace: %w\n\n%s", err, w.Out())
		}

		jj, err := json.MarshalIndent(w.Outputs(), "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal outputs: %w", err)
		}
		fmt.Println("Outputs")
		fmt.Println(string(jj))

		if tfConfirmDelete {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("\n\nPress [enter] to destroy\n")
			_, _ = reader.ReadString('\n')
		}

		fmt.Println("Destroy...")
		if err := w.Destroy(); err != nil {
			return fmt.Errorf("destroy workspace: %w\n\n%s", err, w.Out())
		}

		return nil
	},
}

func init() {
	cmdCloudTf.Flags().StringArrayVarP(&tfProviders, "provider", "p", nil, "tf provider, can be repeated, ie `hashicorp/aws ~> 6.0`")
	cmdCloudTf.Flags().StringArrayVarP(&tfVariables, "var", "v", nil, "input variable, can be repeated, ie `key=\"value\"`, or `key=123`")
	cmdCloudTf.Flags().StringVarP(&tfSource, "source", "s", "", "tf module source, ie `terraform-aws-modules/vpc/aws 6.5.1`")
	cmdCloudTf.Flags().BoolVarP(&tfConfirmDelete, "confirm-delete", "c", false, "ask confirmation before destroy")
	cmdCloud.AddCommand(cmdCloudTf)
	util.MustVoid(cmdCloudTf.MarkFlagRequired("source"))
}
