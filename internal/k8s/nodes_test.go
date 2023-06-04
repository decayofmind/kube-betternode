package k8s

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

func TestListNodes(t *testing.T) {
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "schedulable",
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", "current"),
		},
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
	}

	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "unschedulable",
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", "better"),
			Labels: map[string]string{
				"kubernetes.io/os":                 "linux",
				"node.kubernetes.io/instance-type": "p3.2xlarge",
				"nvidia.com/gpu":                   "true",
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
		},
	}

	clientSet := &fake.Clientset{}
	clientSet.Fake.AddReactor("list", "nodes", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		r := &v1.NodeList{Items: []v1.Node{*node1, *node2}}
		return true, r, nil
	})

	nodes, err := ListNodes(clientSet)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if len(nodes) > 1 {
		t.Fail()
	}
}
