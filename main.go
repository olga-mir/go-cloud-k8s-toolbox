package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// creates the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// get a list of all nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	azNodeMap := buildAzNodeMap(nodes)

	// get a list of all namespaces
	// skipNamespaces := []string{"skip-this", "and-that"}
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// loop over all namespaces
	count := 1000
	skippedDeployments := 0
	unparsableLabels := 0
	for _, namespace := range namespaces.Items {
		count -= 1
		if count == 0 {
			break
		}
		//if contains(skipNamespaces, namespace.ObjectMeta.Name) {
		//	continue
		//}

		// get a list of all deployments in the namespace
		deployments, err := clientset.AppsV1().Deployments(namespace.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		// loop over all deployments in the namespace
		for _, deployment := range deployments.Items {
			if *deployment.Spec.Replicas == 0 {
				skippedDeployments += 1
				continue
			}
			countMap := map[string]int{}

			labelKey, label, err := parseLabels(deployment)

			if err != nil {
				unparsableLabels += 1
				fmt.Printf("%s %s deployment app labels can't be parsed\n", namespace.ObjectMeta.Name, deployment.ObjectMeta.Name)
				continue
			}
			pods, err := clientset.CoreV1().Pods(namespace.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelKey, label)})
			if err != nil {
				panic(err.Error())
			}

			// loop over all pods in the deployment
			for _, pod := range pods.Items {
				// check if the pod is not in a failed state
				if pod.Status.Phase != "Failed" {
					countMap[azNodeMap[pod.Spec.NodeName]] += 1
				}
			}

			// fmt.Printf("%s,%s,%d,%d,%d\n", namespace.GetName(), deployment.GetName(), countMap["a"], countMap["b"], countMap["c"])
			widthName := 50
			widthRes := 20
			fmt.Printf("%s%s%s%s%s\n",
				fmt.Sprintf("%-*s", widthName, namespace.GetName()),
				fmt.Sprintf("%-*s", widthName, deployment.GetName()),
				fmt.Sprintf("%-*s", widthRes, toStars(countMap["a"])),
				fmt.Sprintf("%-*s", widthRes, toStars(countMap["b"])),
				fmt.Sprintf("%-*s", widthRes, toStars(countMap["c"])))
		}
	}
	fmt.Printf("Skipped %d deployments with no replicas\n", skippedDeployments)
	fmt.Printf("Skipped %d deployments with unparsable app labels\n", unparsableLabels)
}

func parseLabels(deployment apps.Deployment) (string, string, error) {
	labelKey := "app"
	label := deployment.ObjectMeta.Labels[labelKey]

	if len(label) == 0 {
		labelKey = "k8s-app"
		label = deployment.ObjectMeta.Labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "app.kubernetes.io/name"
		label = deployment.ObjectMeta.Labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "application"
		label = deployment.ObjectMeta.Labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "app"
		label = deployment.Spec.Template.ObjectMeta.Labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "k8s-app"
		label = deployment.Spec.Template.ObjectMeta.Labels[labelKey]
	}
	if len(label) == 0 {
		return "", "", errors.New("Unable to parse labels")
	}
	return labelKey, label, nil
}

func toStars(num int) string {
	result := ""
	for i := 0; i < num; i++ {
		result += "*"
	}
	return result
}

func buildAzNodeMap(nodes *v1.NodeList) map[string]string {
	// create a map to store nodes by availability zone
	nodeToAzMap := make(map[string]string)

	// loop over all nodes and group them by availability zone
	for _, node := range nodes.Items {
		zone := getNodeZone(node)
		nodeToAzMap[node.ObjectMeta.Name] = zone
	}
	return nodeToAzMap
}

// helper function to get the availability zone of a node
func getNodeZone(node v1.Node) string {
	for key, value := range node.ObjectMeta.Labels {
		if strings.HasPrefix(key, "topology.kubernetes.io/zone") {
			return value[len(value)-1:]
		}
	}

	return "unknown"
}

func contains(arr []string, value string) bool {
	for _, a := range arr {
		if a == value {
			return true
		}
	}
	return false
}
