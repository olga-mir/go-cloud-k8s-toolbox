package k8s

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCmdWorkloadSpread() *cobra.Command {
	var outputFormat string
	cmdWorkloadSpread := &cobra.Command{
		Use:        "spread-by-zone",
		Aliases:    []string{},
		SuggestFor: []string{},

		Short:   "Spread workloads by zone",
		GroupID: "",
		Long:    `TBD - spread by zone (long description TODO))`,
		Example: "",
		//ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		//},
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
			if err := workloadsSpreadByZoneHandler(outputFormat); err != nil {
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

	return cmdWorkloadSpread
}

type PodsSpreadResult struct {
	namespace      string
	controllerName string
	countMap       map[string]int
}

func workloadsSpreadByZoneHandler(outputFormat string) error {
	ctx := context.Background()
	result, uniqueZones, err := workloadsSpreadByZone(ctx)
	if err != nil {
		return err
	}
	output(outputFormat, result, uniqueZones)
	return nil
}

// parse pods by zones and save result in a struct
func workloadsSpreadByZone(ctx context.Context) ([]PodsSpreadResult, map[string]int, error) {
	nodesByZone, err := k8sClient.NodeToZoneMap(ctx)
	if err != nil {
		return nil, nil, err
	}
	namespaces, err := k8sClient.ListNamespaces(ctx)
	if err != nil {
		return nil, nil, err
	}

	// final counts of pods per zone by namespace and controller (deployment or statefulset)
	result := []PodsSpreadResult{}

	// keep track of each zone where pods were found
	uniqueZones := map[string]int{}

	// parse deployments and statefulsets by the namespace.
	for _, namespace := range namespaces {
		deploymentList, err := k8sClient.ListDeployments(ctx, namespace)
		if err != nil || deploymentList == nil {
			return nil, nil, err
		}

		for _, deployment := range deploymentList.Items {
			if *deployment.Spec.Replicas == 0 {
				continue
			}

			countMap := map[string]int{}
			podList, err := k8sClient.ListPodsByLabels(ctx, namespace, deployment.Spec.Selector.MatchLabels)
			if err != nil {
				return nil, nil, err
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
	return result, uniqueZones, nil
}

func output(outputFormat string, result []PodsSpreadResult, uniqueZones map[string]int) {
	if outputFormat == "csv" {
		outputCsv(result, uniqueZones)
	} else if outputFormat == "text" {
		outputText(result, uniqueZones)
	}
}

func outputCsv(result []PodsSpreadResult, uniqueZones map[string]int) error {
	file, err := os.Create("pod-spread-by-zone.csv")
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)

	// get all the keys from the map
	allZones := make([]string, len(uniqueZones))
	i := 0
	for k := range uniqueZones {
		allZones[i] = k
		i++
	}

	err = csvWriter.Write(append([]string{"namespace", "controller"}, allZones...))
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

		err := csvWriter.Write(append([]string{line.namespace, line.controllerName}, zoneCounts...))
		if err != nil {
			return fmt.Errorf("failed to write to output file: %v", err)
		}
	}

	csvWriter.Flush()
	return nil
}

func outputText(result []PodsSpreadResult, uniqueZones map[string]int) error {
	// TODO
	// 	fmt.Printf("CSV,%s,%s,%d,%d,%d\n", namespace, name, countMap["a"], countMap["b"], countMap["c"])
	// 	fmt.Printf("%s%s%s%s%s\n",
	// 		fmt.Sprintf("%-*s", widthName, namespace),
	// 		fmt.Sprintf("%-*s", widthName, name),
	// 		fmt.Sprintf("%-*s", widthRes, toStars(countMap["a"])),
	// 		fmt.Sprintf("%-*s", widthRes, toStars(countMap["b"])),
	// 		fmt.Sprintf("%-*s", widthRes, toStars(countMap["c"])))
	return nil
}

// func toStars(num int) string {
// result := ""
// for i := 0; i < num; i++ {
// result += "*"
// }
// return result
// }
