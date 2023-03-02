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
	// "k8s.io/client-go/rest"
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

	// get a list of all namespaces
	namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// loop over all namespaces
	for _, namespace := range namespaces.Items {
		fmt.Printf("Namespace: %s\n", namespace.ObjectMeta.Name)

		// get a list of all deployments in the namespace
		deployments, err := clientset.AppsV1().Deployments(namespace.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		// loop over all deployments in the namespace
		for _, deployment := range deployments.Items {
			fmt.Printf("\tDeployment: %s\n", deployment.ObjectMeta.Name)

			// get a list of all pods in the deployment
			pods, err := clientset.CoreV1().Pods(namespace.ObjectMeta.Name).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", deployment.ObjectMeta.Name)})
			if err != nil {
				panic(err.Error())
			}

			// loop over all pods in the deployment
			for _, pod := range pods.Items {
				// check if the pod is not in a failed state
				if pod.Status.Phase != "Failed" {
					fmt.Printf("\t\tPod: %s\n", pod.ObjectMeta.Name)
				}
			}
		}
	}

	// get a list of all nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// create a map to store nodes by availability zone
	nodeMap := make(map[string][]string)

	// loop over all nodes and group them by availability zone
	for _, node := range nodes.Items {
		zone := getNodeZone(node)

		if _, ok := nodeMap[zone]; !ok {
			nodeMap[zone] = []string{}
		}

		nodeMap[zone] = append(nodeMap[zone], node.ObjectMeta.Name)
	}

	// print out the nodes by availability zone
	for zone, nodes := range nodeMap {
		fmt.Printf("Availability Zone: %s\n", zone)

		for _, node := range nodes {
			fmt.Printf("\tNode: %s\n", node)
		}
	}
}

// helper function to get the availability zone of a node

func getNodeZone(node v1.Node) string {
	for _, label := range node.ObjectMeta.Labels {
		if strings.HasPrefix(label, "failure-domain.beta.kubernetes.io/zone=") {
			return strings.TrimPrefix(label, "failure-domain.beta.kubernetes.io/zone=")
		}
	}

	return "unknown"
}
