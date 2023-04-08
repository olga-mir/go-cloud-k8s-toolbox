package k8s

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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

// helper function to get the availability zone of a node
func getNodeZone(node v1.Node) string {
	for key, value := range node.ObjectMeta.Labels {
		if strings.HasPrefix(key, "topology.kubernetes.io/zone") {
			// only a lettter a, b, etc:  value[len(value)-1:]
			return value
		}
	}
	return "unknown"
}

// NodeToZoneMap returns a map that maps node name to AZ it runs in
func (c *Client) NodeToZoneMap(ctx context.Context) (map[string]string, error) {

	nodeToZoneMap := make(map[string]string)

	nodeList, err := c.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	for _, node := range nodeList.Items {
		nodeToZoneMap[node.ObjectMeta.Name] = getNodeZone(node)
	}

	return nodeToZoneMap, nil
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

func labelsToSelector(labels map[string]string) string {
	selector := []string{}

	for k, v := range labels {
		selector = append(selector, k+"="+v)
	}

	return strings.Join(selector, ",")
}

func (c *Client) ListPodsByLabels(ctx context.Context, clientset kubernetes.Interface, ns string, labels map[string]string) (*v1.PodList, error) {
	selector := labelsToSelector(labels)
	podList, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return podList, nil
}

// for each deployment in the namespace, get the labels map which is union of metadata.labels and spec.template.metadata.labels
// TODO - this is too specific for `pkg/k8s`. Move it to `cmd` or `internal` package
/*func (c *Client) LabelsOfNonEmptyDeployments(ctx context.Context, ns string) (map[string]map[string]string, error) {
	result := map[string]map[string]string{}

	deployments, err := c.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, deployment := range deployments.Items {
		// l1 := deployment.ObjectMeta.Labels
		// l2 := deployment.Spec.Template.ObjectMeta.Labels
		// merge the two maps
		//for k, v := range l2 {
		//	l1[k] = v
		//}
		// result[deployment.ObjectMeta.Name] = l1

		// probably it is possible to drop all of the label parsing code from original implementaion and
		// just use the labels from the deployment.Spec.Template.ObjectMeta.Labels.
		result[deployment.ObjectMeta.Name] = deployment.Spec.Template.ObjectMeta.Labels

		// fmt.Printf("deployment: %s, labels: %v and %v", deployment.Name, l1, l2)
	}
	return result, nil
}
*/

// return list of deployments with non-zero replicas

// gererate function that returns a list of v1.Deployment objects
func (c *Client) ListDeployments(ctx context.Context, ns string) (*appsv1.DeploymentList, error) {
	deployments, err := c.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deployments, nil
}
