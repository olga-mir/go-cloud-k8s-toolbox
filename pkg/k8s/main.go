package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Clientset kubernetes.Interface
	// Config            *rest.Config
	// RawConfig         clientcmdapi.Config
	// contextName       string
}

// NewClient returns a new client and error
func NewClient(kubeconfig string) (*Client, error) {
	client := &Client{}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client.Clientset = clientset
	return client, nil
}

// ListNodes returns a list of nodes and error
func (c *Client) ListNodes(ctx context.Context) ([]string, error) {
	//create a new list of nodes
	nodes := []string{}
	// list all nodes
	nodeList, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nodes, err
	}
	// add all nodes to the list
	for _, node := range nodeList.Items {
		nodes = append(nodes, node.Name)
	}
	return nodes, nil
}

// ListNodesByZone returns a map of nodes by zone and error
func (c *Client) ListNodesByZone(ctx context.Context) (map[string][]string, error) {

	nodesByZone := make(map[string][]string)

	nodeList, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodeList.Items {
		zone := node.Labels["topology.kubernetes.io/zone"]
		nodesByZone[zone] = append(nodesByZone[zone], node.Name)
	}
	return nodesByZone, nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces := []string{}
	namespaceList, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return namespaces, err
	}
	for _, namespace := range namespaceList.Items {
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

// for each deployment in the namespace, get the labels map which is union of metadata.labels and spec.template.metadata.labels
// TODO - this is too specific for `pkg/k8s`. Move it to `cmd` or `internal` package
func (c *Client) LabelsOfNonEmptyDeployments(ctx context.Context, ns string) (map[string]map[string]string, error) {
	result := map[string]map[string]string{}

	deployments, err := c.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, deployment := range deployments.Items {
		l1 := deployment.ObjectMeta.Labels
		l2 := deployment.Spec.Template.ObjectMeta.Labels

		// merge the two maps
		for k, v := range l2 {
			l1[k] = v
		}
		result[deployment.ObjectMeta.Name] = l1

		fmt.Printf("deployment: %s, labels: %v and %v", deployment.Name, l1, l2)
	}
	return result, nil
}
