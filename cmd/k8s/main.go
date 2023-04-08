package k8s

import (
	"context"
	"fmt"

	client "github.com/olga-mir/go-cloud-k8s-toolbox/pkg/k8s"
)

func main() {
	client, err := client.NewClient("")
	if err != nil {
		panic(err)
	}

	nodesByZone, err := client.ListNodesByZone(context.Background())
	if err != nil {
		panic(err)
	}

	// print the map
	for zone, nodes := range nodesByZone {
		fmt.Printf("%s: %s", zone, nodes)
	}
}
