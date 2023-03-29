package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var widthName = 50
var widthRes = 20

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
		//if !contains(onlyNamespaces, namespace.ObjectMeta.Name) {
		//	continue
		//}
		ns := namespace.ObjectMeta.Name

		// get a list of all deployments in the namespace
		deployments, err := clientset.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
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
			labelKey, label, err := parseLabels(deployment.ObjectMeta.Labels, deployment.Spec.Template.ObjectMeta.Labels)
			if err != nil {
				unparsableLabels += 1
				fmt.Printf("%s %s deployment app labels can't be parsed\n", ns, deployment.ObjectMeta.Name)
				continue
			}
			pods, err := clientset.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelKey, label)})
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

			output(deployment.GetName(), namespace.GetName(), countMap)
		}

		statefulSets, err := clientset.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, sts := range statefulSets.Items {
			countMap := map[string]int{}
			labelKey, label, err := parseLabels(sts.ObjectMeta.Labels, sts.Spec.Template.ObjectMeta.Labels)
			if err != nil {
				// unparsableLabels += 1
				fmt.Printf("%s %s statefullSet app labels can't be parsed\n", namespace.ObjectMeta.Name, sts.ObjectMeta.Name)
				continue
			}

			stsPods, err := clientset.CoreV1().Pods(namespace.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", labelKey, label)})
			if err != nil {
				panic(err.Error())
			}

			for _, pod := range stsPods.Items {
				// check if the pod is not in a failed state
				if pod.Status.Phase != "Failed" {
					countMap[azNodeMap[pod.Spec.NodeName]] += 1
				}
			}
			output(sts.GetName(), namespace.GetName(), countMap)
		}
	}
	fmt.Printf("Skipped %d deployments with no replicas\n", skippedDeployments)
	fmt.Printf("Skipped %d deployments with unparsable app labels\n", unparsableLabels)
}

func output(name string, namespace string, countMap map[string]int) {
	fmt.Printf("CSV,%s,%s,%d,%d,%d\n", namespace, name, countMap["a"], countMap["b"], countMap["c"])
	fmt.Printf("%s%s%s%s%s\n",
		fmt.Sprintf("%-*s", widthName, namespace),
		fmt.Sprintf("%-*s", widthName, name),
		fmt.Sprintf("%-*s", widthRes, toStars(countMap["a"])),
		fmt.Sprintf("%-*s", widthRes, toStars(countMap["b"])),
		fmt.Sprintf("%-*s", widthRes, toStars(countMap["c"])))
}

func parseLabels(labels map[string]string, specLabels map[string]string) (string, string, error) {
	labelKey := "app"
	label := labels[labelKey]

	if len(label) == 0 {
		labelKey = "k8s-app"
		label = labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "app.kubernetes.io/name"
		label = labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "application"
		label = labels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "app"
		label = specLabels[labelKey]
	}
	if len(label) == 0 {
		labelKey = "k8s-app"
		label = specLabels[labelKey]
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
