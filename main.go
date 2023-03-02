package main

import (
	"context"
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
	skipNamespaces := []string{"skip-this", "and-that"}
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// loop over all namespaces
	count := 80
	skippedDeployments := 0
	for _, namespace := range namespaces.Items {
		count -= 1
		if count == 0 {
			break
		}
		if contains(skipNamespaces, namespace.ObjectMeta.Name) {
			continue
		}

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

			// get a list of all pods in the deployment
			pods, err := clientset.CoreV1().Pods(namespace.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", deployment.ObjectMeta.Name)})
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
			fmt.Printf("%s,%s,%d,%d,%d\n", namespace.GetName(), deployment.GetName(), countMap["a"], countMap["b"], countMap["c"])
			width := 20
			fmt.Printf("%s%s%s%s%s\n",
				fmt.Sprintf("%-*s", width, namespace.GetName()),
				fmt.Sprintf("%-*s", width, deployment.GetName()),
				fmt.Sprintf("%-*s", width, toStars(countMap["a"])),
				fmt.Sprintf("%-*s", width, toStars(countMap["b"])),
				fmt.Sprintf("%-*s", width, toStars(countMap["c"])))
		}
	}
	fmt.Printf("Skipped %d deployments with no replicas\n", skippedDeployments)
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
