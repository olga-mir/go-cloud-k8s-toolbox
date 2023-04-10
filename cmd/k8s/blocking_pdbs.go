package k8s

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	client "github.com/olga-mir/go-cloud-k8s-toolbox/pkg/k8s"
)

func newCmdBlockingPDBs() *cobra.Command {

	validArgs := []string{"blocking-pdbs"}

	cmdBlockingPDBs := &cobra.Command{
		Use:   fmt.Sprintf("blocking-pdbs [flags] %s", validArgs),
		Short: "Find blocking PDBs",
		Long:  `TBD - find blocking PDBs (long description TODO))`,

		ValidArgs: validArgs,
		Args:      cobra.MatchAll(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing subcommand, valid subcommands are: %s", validArgs)
			}
			var kubeconfig string
			cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
			if kubeconfig == "" {
				if home := homedir.HomeDir(); home != "" {
					kubeconfig = filepath.Join(home, ".kube", "config")
				}
			}
			// client, err := client.NewClient(kubeconfig)
			_, err := client.NewClient(kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to create k8s client: %v", err)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			var outputFormat string
			cmd.Flags().StringVar(&outputFormat, "output", "", "output format (csv or text), when not specified, output is printed to stdout")
			fmt.Printf("outputFormat: %s, args: %s", outputFormat, args)
			//if err := do_stuff; err != nil {
			//	return err
			//}
			return nil
		},
	}

	return cmdBlockingPDBs
}
