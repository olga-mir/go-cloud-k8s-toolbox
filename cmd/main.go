package main

import (
	"fmt"

	"github.com/olga-mir/go-cloud-k8s-toolbox/cmd/k8s"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-cloud-k8s-toolbox",
	Short: "Helper funtions to work with cloud and/or kubernetes. Your one-liners and bash scripts but better",
	Long:  `TBD`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running root command")
	},
}

func main() {
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.toolbox.yaml)")
	rootCmd.AddCommand(k8s.NewK8sCmd())
	rootCmd.Execute()
}
