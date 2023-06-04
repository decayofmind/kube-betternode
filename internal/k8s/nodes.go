package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func ListNodes(client kubernetes.Interface) ([]*v1.Node, error) {
	nodeList, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []*v1.Node{}, err
	}

	nodes := make([]*v1.Node, 0)
	for i := range nodeList.Items {
		node := &nodeList.Items[i]

		if node.Spec.Unschedulable {
			klog.V(4).InfoS("Skipping unschedulable node", "node", klog.KObj(node))
			continue
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}
