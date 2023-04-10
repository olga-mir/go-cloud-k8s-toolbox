package k8s

import (
	"github.com/spf13/cobra"
)

// this function registers all the commands available at "wizardry k8s" level
func NewK8sCmd() *cobra.Command {
	// validArgs := []string{"spread-by-zone", "blocking-pdbs"}
	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "Host all subcommands related to k8s",
		Long:  `TBD`,
		// ValidArgs: validArgs,
		//Run: func(cmd *cobra.Command, args []string) {
		//	fmt.Println("RUN - k8s subcommand")
		//},
	}
	cmd.AddCommand(newCmdWorkloadSpread())
	cmd.AddCommand(newCmdBlockingPDBs())
	return cmd
}
