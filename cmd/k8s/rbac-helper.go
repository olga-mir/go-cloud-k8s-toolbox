package k8s

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func newCmdRBACComposer() *cobra.Command {
	cmdRBACComposer := &cobra.Command{
		Use:        "rbac-composer",
		Aliases:    []string{},
		SuggestFor: []string{},

		Short:   "Create RBAC objects",
		GroupID: "",
		Long:    ``,
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
			if err := RBACCompose(); err != nil {
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

	// cmdRBACComposer.Flags().StringVar(&varname, "name", "", "desc")

	return cmdRBACComposer
}

type ClusterRole struct {
	Kind       string     `yaml:"kind"`
	APIVersion string     `yaml:"apiVersion"`
	Metadata   Metadata   `yaml:"metadata"`
	Rules      []RoleRule `yaml:"rules"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type RoleRule struct {
	APIGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

func RBACCompose() error {
	// input file is output of kubectl api-resources but with the first line removed and
	// removed the SHORTNAMES column manually
	file, err := os.Open("api-resources.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Create a new ClusterRole
	role := ClusterRole{
		Kind:       "ClusterRole",
		APIVersion: "rbac.authorization.k8s.io/v1",
		Metadata: Metadata{
			Name: "no-secrets-access",
		},
		Rules: []RoleRule{},
	}

	// Rules for core pods and RBAC will be hardcoded separately and should not be auto-imported from api-resources
	sensitiveApiGroups := []string{"rbac.authorization.k8s.io"}
	sensitiveCoreResources := []string{"pods", "secrets"}

	prevApiGroup := "core" // TODO - assumes first line is core group
	resources := []string{}
	apigroup := ""
	for scanner.Scan() {
		// each line is from "k api-resources" but shortnames column should be removed manually from the input file
		//  flowschemas                                    flowcontrol.apiserver.k8s.io/v1beta2   false        FlowSchema
		line := scanner.Text()

		fields := strings.Fields(line)
		groups := strings.Split(fields[1], "/")

		if len(groups) > 1 {
			apigroup = groups[0]
		} else {
			apigroup = "core"
		}

		if apigroup != prevApiGroup {
			if !contains(sensitiveApiGroups, prevApiGroup) {
				rule := RoleRule{
					APIGroups: []string{prevApiGroup},
					Resources: resources,
					Verbs:     []string{"*"},
				}
				role.Rules = append(role.Rules, rule)
			}
			resources = []string{}
			prevApiGroup = apigroup
		}

		if !(prevApiGroup == "core" && contains(sensitiveCoreResources, fields[0])) {
			resources = append(resources, fields[0])
		}
	}

	podRule := RoleRule{
		APIGroups: []string{"core"},
		Resources: []string{"pods"},
		Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
	}
	role.Rules = append(role.Rules, podRule)

	rbacRule := RoleRule{
		APIGroups: []string{"rbac.authorization.k8s.io"},
		Resources: []string{"clusterrolebindings", "clusterroles", "rolebindings", "roles"},
		Verbs:     []string{"get", "list"},
	}
	role.Rules = append(role.Rules, rbacRule)

	if err := scanner.Err(); err != nil {
		return err
	}

	// Convert the ClusterRole to YAML
	roleYAML, err := yaml.Marshal(role)
	if err != nil {
		return err
	}

	// Write the YAML to a file
	err = ioutil.WriteFile("clusterrole.yaml", roleYAML, 0644)
	if err != nil {
		return err
	}

	fmt.Println("ClusterRole written to clusterrole.yaml")
	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
