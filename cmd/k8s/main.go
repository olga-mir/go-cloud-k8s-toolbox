package k8s

import (
	"fmt"
	"path/filepath"

	client "github.com/olga-mir/go-cloud-k8s-toolbox/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

type K8sCmd struct {
	client *client.Client
	cmd    *cobra.Command
}

var k8sCmd K8sCmd

// this function registers all the commands available at "wizardry k8s" level
func NewK8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Helper functions to work with k8s clusters",
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("RUN - k8s subcommand, args: %v\n", args)
		},
	}

	var kubeconfig string
	cmd.PersistentFlags().String("kubeconfig", "", "kubeconfig file (default is $HOME/.kube/config.yaml)")
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	c, err := client.NewClient(kubeconfig)
	if err != nil {
		fmt.Printf("failed to create k8s client: %v", err)
		return nil
	}

	k8sCmd.client = c
	k8sCmd.cmd = cmd

	cmd.AddCommand(newCmdWorkloadSpread())
	cmd.AddCommand(newCmdBlockingPDBs())
	cmd.AddCommand(newCmdRBACComposer())
	return cmd
}
