package k8s

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

func TestListPods(t *testing.T) {
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			UID:       types.UID("pod"),
		},
		Spec: v1.PodSpec{
			NodeName: "node",
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

	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
			UID:       types.UID("pod"),
		},
		Spec: v1.PodSpec{
			NodeName: "node",
			Containers: []v1.Container{
				{
					Name:  "main",
					Image: "busybox",
				},
			},
			NodeSelector: map[string]string{
				"kubernetes.io/os": "linux",
			},
		},
		Status: v1.PodStatus{
			PodIP:  "1.2.3.5",
			HostIP: "2.3.4.6",
			Phase:  "Running",
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}
	clientSet := &fake.Clientset{}
	clientSet.Fake.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		r := &v1.PodList{Items: []v1.Pod{*pod1, *pod2}}
		return true, r, nil
	})

	pods, err := ListPods(clientSet)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if len(pods) > 1 {
		t.Fail()
	}
}
