package k8s

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	client "github.com/olga-mir/go-cloud-k8s-toolbox/pkg/k8s"
)

func NewCmdWorkloadSpread(ctx context.Context) (*cobra.Command, error) {

	validArgs := []string{"spread-by-zone", "output"}

	cmdWorkloadSpread := &cobra.Command{
		Use:   fmt.Sprintf("k8s [flags] %s", validArgs),
		Short: "Aux tools to work with k8s",
		Long:  `TBD`,

		ValidArgs: validArgs,
		Args:      cobra.MatchAll(),
		RunE: func(cmd *cobra.Command, args []string) error {
			var outputFormat string
			cmd.Flags().StringVar(&outputFormat, "output", "", "output format (csv or text), when not specified, output is printed to stdout")
			if err := workloadsSpreadByZone(outputFormat); err != nil {
				return err
			}
			return nil
		},
	}

	return cmdWorkloadSpread, nil
}

func workloadsSpreadByZone(outputFormat string) error {
	ctx := context.Background()
	client, err := client.NewClient("")
	if err != nil {
		panic(err)
	}
	fmt.Printf("output: %s", outputFormat)

	nodesByZone, err := client.ListNodesByZone(ctx)
	if err != nil {
		return err
	}

	// print the map
	for zone, nodes := range nodesByZone {
		fmt.Printf("%s: %s", zone, nodes)
	}

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		return err
	}

	// parse deployments and statefulsets by the namespace.
	// TODO - reduce number of calls by listing all deployments and statefulsets in one (paged) call
	for _, namespace := range namespaces {
		labels := client.LabelsOfNonEmptyDeployments(ctx, namespace)
	}

	return nil
}
