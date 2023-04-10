package k8s

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	client "github.com/olga-mir/go-cloud-k8s-toolbox/pkg/k8s"
)

type Cmd struct {
	client client.Client
	ctx    context.Context
}

func newCmdWorkloadSpread() *cobra.Command {

	validArgs := []string{}
	var c *Cmd

	cmdWorkloadSpread := &cobra.Command{
		Use:   "spread-by-zone", // fmt.Sprintf("spread-by-zone [flags] %s", validArgs),
		Short: "Spread workloads by zone",
		Long:  `TBD - spread by zone (long description TODO))`,

		/// ValidArgs: validArgs,
		Args: cobra.MatchAll(),
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
			client, err := client.NewClient(kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to create k8s client: %v", err)
			}
			c = newCmd(*client)
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO - flags parsing it not working (needs another level of cmd.AddCommand)
			var outputFormat string
			cmd.Flags().StringVar(&outputFormat, "output", "", "output format (csv or text), when not specified, output is printed to stdout")
			fmt.Printf("outputFormat: %s, args: %s", outputFormat, args)
			if err := c.workloadsSpreadByZone(outputFormat); err != nil {
				return err
			}
			return nil
		},
	}

	return cmdWorkloadSpread
}

func newCmd(client client.Client) *Cmd {
	return &Cmd{
		client: client,
		ctx:    context.Background(),
	}
}

type PodsSpreadResult struct {
	namespace      string
	controllerName string
	countMap       map[string]int
}

func (c *Cmd) workloadsSpreadByZone(outputFormat string) error {

	nodesByZone, err := c.client.NodeToZoneMap(c.ctx)
	if err != nil {
		return err
	}

	namespaces, err := c.client.ListNamespaces(c.ctx)
	if err != nil {
		return err
	}

	// final counts of pods per zone by namespace and controller (deployment or statefulset)
	result := []PodsSpreadResult{}

	// keep track of each zone where pods were found
	uniqueZones := map[string]int{}

	// parse deployments and statefulsets by the namespace.
	// TODO - reduce number of calls by listing all deployments and statefulsets in one (paged) call
	for _, namespace := range namespaces {
		deploymentList, err := c.client.ListDeployments(c.ctx, namespace)
		if err != nil || deploymentList == nil {
			return err
		}

		for _, deployment := range deploymentList.Items {
			if *deployment.Spec.Replicas == 0 {
				continue
			}

			countMap := map[string]int{}
			podList, err := c.client.ListPodsByLabels(c.ctx, namespace, deployment.Spec.Selector.MatchLabels)
			if err != nil {
				return err
			}
			for _, pod := range podList.Items {
				if pod.Status.Phase != "Failed" {
					zone := nodesByZone[pod.Spec.NodeName]
					uniqueZones[zone] += 1
					countMap[zone] += 1
				}
			}
			result = append(result, PodsSpreadResult{
				namespace:      namespace,
				controllerName: deployment.ObjectMeta.Name,
				countMap:       countMap,
			})
		}
		// TODO - the same for statefulsets
	}
	output(outputFormat, result, uniqueZones)

	return nil
}

func output(outputFormat string, result []PodsSpreadResult, uniqueZones map[string]int) error {
	file, err := os.Create("pod-spread-by-zone.csv")
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	// get all the keys from the map
	allZones := make([]string, len(uniqueZones))
	i := 0 // this approach is a bit more efficient than using append
	for k := range uniqueZones {
		allZones[i] = k
		i++
	}

	err = writer.Write(append([]string{"namespace", "controller"}, allZones...))
	if err != nil {
		return fmt.Errorf("failed to write header to output file: %v", err)
	}

	for _, line := range result {
		zoneCounts := []string{}

		// assuming iteration order is stable
		for _, zone := range allZones {
			if _, ok := line.countMap[zone]; !ok {
				zoneCounts = append(zoneCounts, "0")
			} else {
				zoneCounts = append(zoneCounts, fmt.Sprintf("%d", line.countMap[zone]))
			}
		}

		err := writer.Write(append([]string{line.namespace, line.controllerName}, zoneCounts...))
		if err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
	}

	writer.Flush()
	return nil
}
