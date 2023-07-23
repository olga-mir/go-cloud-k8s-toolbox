package k8s

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCmdWorkloadSpread() *cobra.Command {

	// 'csv' or 'text' for representing pods as wildcards per zone.
	var outputFormat string

	// short - if true only output workloads that have significant disbalance
	var short bool

	cmdWorkloadSpread := &cobra.Command{
		Use:        "spread-by-zone",
		Aliases:    []string{},
		SuggestFor: []string{},

		Short:                  "Spread workloads by zone",
		GroupID:                "",
		Long:                   `TBD - spread by zone (long description TODO))`,
		Example:                "",
		Args:                   cobra.MatchAll(),
		ArgAliases:             []string{},
		BashCompletionFunction: "",
		Deprecated:             "",
		Annotations:            map[string]string{},
		Version:                "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := k8sCmd.workloadsSpreadByZoneHandler(outputFormat); err != nil {
				return err
			}
			return nil
		},
		PostRun: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
		},
		FParseErrWhitelist:         cobra.FParseErrWhitelist{},
		CompletionOptions:          cobra.CompletionOptions{},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: 0,
	}

	cmdWorkloadSpread.Flags().StringVar(&outputFormat, "output", "", "output format (csv or text), when not specified, output is printed to stdout")
	cmdWorkloadSpread.Flags().BoolVar(&short, "short", false, "only output workloads with uneven spread")

	return cmdWorkloadSpread
}

type WorkloadSpread struct {
	namespace      string
	controllerName string
	countMap       map[string]int
}

type Result struct {
	// one line for each top level controller describing its pods spread
	spread []WorkloadSpread

	// number of zones and their names are detected dynamically and stored in this field
	zoneNames []string
}

func (c *K8sCmd) workloadsSpreadByZoneHandler(outputFormat string) error {
	ctx := context.Background()
	result, err := c.workloadsSpreadByZone(ctx)
	if err != nil {
		return err
	}
	output(outputFormat, *result)
	return nil
}

// TODO - remove result passing around as pointer. a field on reciever instead?
func (c *K8sCmd) workloadsSpreadByZone(ctx context.Context) (*Result, error) {
	nodesByZone, err := c.client.NodeToZoneMap(ctx)
	if err != nil {
		return nil, err
	}
	namespaces, err := c.client.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	// final counts of pods per zone by namespace and controller (deployment or statefulset)
	result := Result{
		spread:    []WorkloadSpread{},
		zoneNames: []string{},
	}

	// keep track of each zone where pods were found
	uniqueZones := map[string]int{}

	// parse deployments and statefulsets by the namespace.
	for _, namespace := range namespaces {
		deploymentList, err := c.client.ListDeployments(ctx, namespace)
		if err != nil || deploymentList == nil {
			return nil, err
		}

		for _, deployment := range deploymentList.Items {
			if *deployment.Spec.Replicas == 0 {
				continue
			}

			countMap := map[string]int{}
			podList, err := c.client.ListPodsByLabels(ctx, namespace, deployment.Spec.Selector.MatchLabels)
			if err != nil {
				return nil, err
			}
			for _, pod := range podList.Items {
				// TODO - is this needed? other values to check for?
				if pod.Status.Phase != "Failed" {
					zone := nodesByZone[pod.Spec.NodeName]
					uniqueZones[zone] += 1
					countMap[zone] += 1
				}
			}
			result.spread = append(result.spread, WorkloadSpread{
				namespace:      namespace,
				controllerName: deployment.ObjectMeta.Name,
				countMap:       countMap,
			})
		}
		// get all the keys from the map
		result.zoneNames = make([]string, len(uniqueZones))
		i := 0
		for k := range uniqueZones {
			result.zoneNames[i] = k
			i++
		}

		// for each workload set zero for each zone where its pods don't exist
		normaliseResult(&result)

		// TODO - the same for statefulsets
	}
	return &result, nil
}

// make sure each workload has all the zones in the map with 0 if there are no pods of this workload
func normaliseResult(result *Result) {
	for _, workload := range result.spread {
		for _, zone := range result.zoneNames {
			if _, ok := workload.countMap[zone]; !ok {
				workload.countMap[zone] = 0
			}
		}
	}
}

// workload spread represents one deployment/sts pod count per zone
// {a: 2, b: 4} etc
//func calcDisbalance(WorkloadSpread ws) {
//
//}

func output(outputFormat string, result Result) {
	if outputFormat == "csv" {
		outputCsv(result)
	} else if outputFormat == "text" {
		outputText(result)
	}
}

func outputCsv(result Result) error {
	file, err := os.Create("pod-spread-by-zone.csv")
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)

	err = csvWriter.Write(append([]string{"namespace", "controller"}, result.zoneNames...))
	if err != nil {
		return fmt.Errorf("failed to write header to output file: %v", err)
	}

	for _, line := range result.spread {
		zoneCounts := []string{}

		// assuming iteration order is stable
		for _, zone := range result.zoneNames {
			zoneCounts = append(zoneCounts, fmt.Sprintf("%d", line.countMap[zone]))
		}

		err := csvWriter.Write(append([]string{line.namespace, line.controllerName}, zoneCounts...))
		if err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
	}

	csvWriter.Flush()
	return nil
}

func outputText(result Result) error {
	// TODO - for now drop to stdout
	var widthName = 50
	var widthRes = 20
	for _, line := range result.spread {
		outputCounts := ""

		// assuming iteration order is stable
		for _, zone := range result.zoneNames {
			outputCounts = fmt.Sprintf("%-*s", widthRes, toStars(line.countMap[zone]))

			fmt.Printf("%s%s%s\n",
				fmt.Sprintf("%-*s", widthName, line.namespace),
				fmt.Sprintf("%-*s", widthName, line.controllerName),
				fmt.Sprintf("%-*s", widthRes, outputCounts))
		}
	}

	return nil
}

// converts a number to that number of wildcards, 3 -> `***`
func toStars(num int) string {
	result := ""
	for i := 0; i < num; i++ {
		result += "*"
	}
	return result
}
