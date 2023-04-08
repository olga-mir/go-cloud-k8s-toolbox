package k8s

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	client "github.com/olga-mir/go-cloud-k8s-toolbox/pkg/k8s"
)

func NewCmdWorkloadSpread(ctx context.Context) *cobra.Command {

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

	return cmdWorkloadSpread
}

type PodsSpreadResult struct {
	namespace      string
	controllerName string
	countMap       map[string]int
}

func workloadsSpreadByZone(outputFormat string) error {
	ctx := context.Background()

	client, err := client.NewClient("")
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %v", err)
	}
	fmt.Printf("output: %s", outputFormat)

	nodesByZone, err := client.NodeToZoneMap(ctx)
	if err != nil {
		return err
	}

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		return err
	}

	result := []PodsSpreadResult{}

	// parse deployments and statefulsets by the namespace.
	// TODO - reduce number of calls by listing all deployments and statefulsets in one (paged) call
	for _, namespace := range namespaces {
		deploymentList, err := client.ListDeployments(ctx, namespace)
		if err != nil || deploymentList == nil {
			return err
		}

		for _, deployment := range deploymentList.Items {
			if *deployment.Spec.Replicas == 0 {
				continue
			}

			countMap := map[string]int{}
			podList, err := client.ListPodsByLabels(ctx, client.Clientset, namespace, deployment.Spec.Template.ObjectMeta.Labels)
			if err != nil {
				return err
			}
			for _, pod := range podList.Items {
				if pod.Status.Phase != "Failed" {
					countMap[nodesByZone[pod.Spec.NodeName]] += 1
				}
			}
			result = append(result, PodsSpreadResult{
				namespace:      namespace,
				controllerName: deployment.ObjectMeta.Name,
				countMap:       countMap,
			})
		}
		// TODO - the same for statefulsets
		output(outputFormat, result)
	}

	return nil
}

func output(outputFormat string, result []PodsSpreadResult) {
	// TODO - implement output to stdout and to csv file
}
