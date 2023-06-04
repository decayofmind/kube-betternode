package controller

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestFindBetterNode(t *testing.T) {
	nodeCurrent := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "current",
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", "current"),
			Labels: map[string]string{
				"kubernetes.io/os":                 "linux",
				"node.kubernetes.io/instance-type": "c6a.2xlarge",
			},
		},
	}

	nodeBetter := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "better",
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", "better"),
			Labels: map[string]string{
				"kubernetes.io/os":                 "linux",
				"node.kubernetes.io/instance-type": "p3.2xlarge",
				"nvidia.com/gpu":                   "true",
			},
		},
	}

	nodeWrong1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "wrong1",
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", "better"),
			Labels: map[string]string{
				"kubernetes.io/os":                 "windows",
				"node.kubernetes.io/instance-type": "p3.2xlarge",
				"nvidia.com/gpu":                   "true",
			},
		},
	}

	nodeWrong2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:     "wrong2",
			SelfLink: fmt.Sprintf("/api/v1/nodes/%s", "better"),
			Labels: map[string]string{
				"kubernetes.io/os":                 "linux",
				"node.kubernetes.io/instance-type": "p3.8xlarge",
				"nvidia.com/gpu":                   "true",
			},
		},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "default",
			UID:       types.UID("pod"),
		},
		Spec: v1.PodSpec{
			NodeName: nodeCurrent.Name,
			Containers: []v1.Container{
				{
					Name:  "main",
					Image: "busybox",
				},
			},
			NodeSelector: map[string]string{
				"kubernetes.io/os": "linux",
			},
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "node.kubernetes.io/instance-type",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"p3.2xlarge", "c6a.2xlarge"},
									},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 100,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "nvidia.com/gpu",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"true"},
									},
								},
							},
						},
						{
							Weight: 0,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "kubernetes.io/arch",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"amd64", "intel"},
									},
								},
							},
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			PodIP:  "1.2.3.4",
			HostIP: "2.3.4.5",
			Phase:  "Running",
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}

	newNode1, _, _ := FindBetterNode(pod, 0, 0, []*v1.Node{nodeCurrent, nodeBetter, nodeWrong1, nodeWrong2})

	if newNode1 != nil && newNode1.Name != "better" {
		t.Error(newNode1)
	}

	newNode2, _, _ := FindBetterNode(pod, 0, 0, []*v1.Node{nodeCurrent, nodeWrong1, nodeWrong2})
	if newNode2 != nil {
		t.Error(newNode2)
	}
}
