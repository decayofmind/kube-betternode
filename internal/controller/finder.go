package controller

import (
	"github.com/decayofmind/kube-better-node/internal/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/noderesources"
)

func FindBetterNode(client kubernetes.Interface, pod *v1.Pod, curScore int, tolerance int, nodes []*v1.Node) (*v1.Node, int, error) {
	for _, node := range nodes {
		// Skip nodes that do not match the Pod's NodeSelector
		if len(pod.Spec.NodeSelector) > 0 {
			nodeSelector := labels.SelectorFromSet(pod.Spec.NodeSelector)
			if !nodeSelector.Matches(labels.Set(node.Labels)) {
				klog.V(4).InfoS("Skipping node that does not match NodeSelector", "node", klog.KObj(node))
				continue
			}
		}

		// Skip nodes that do not match the Pod's required nodeAffinity
		nodeAffinity := pod.Spec.Affinity.NodeAffinity
		if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			match, err := helpers.MatchNodeSelectorTerms(node, nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
			if err != nil {
				return nil, 0, err
			}

			if !match {
				klog.V(4).InfoS("Skipping node that does not match required nodeAffinity", "node", klog.KObj(node))
				continue
			}
		}

		// Skip nodes that doesn't have the required resources
		podsOnNode, err := k8s.ListPodsOnNode(client, node.ObjectMeta.Name)
		if err != nil {
			panic(err.Error())
		}
		podsWithoutLowPriority := make([]*v1.Pod, 0)
		for _, podNode := range podsOnNode {
			if *podNode.Spec.Priority >= *pod.Spec.Priority {
				podsWithoutLowPriority = append(podsWithoutLowPriority, podNode)
			}
		}
		nodeInfo := framework.NewNodeInfo(podsWithoutLowPriority...)
		nodeInfo.SetNode(node)
		insufficientResources := noderesources.Fits(pod, nodeInfo)
		if len(insufficientResources) != 0 {
			klog.V(4).InfoS("Skipping node that does not match required resources", "node", klog.KObj(node))
			continue
		}

		terms, err := nodeaffinity.NewPreferredSchedulingTerms(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
		if err != nil {
			return nil, 0, err
		}

		score := int(terms.Score(node))

		if (score - tolerance) > curScore {
			return node, score, nil
		}
	}

	return nil, 0, nil
}
