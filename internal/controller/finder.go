package controller

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
	"k8s.io/klog/v2"
)

func FindBetterNode(pod *v1.Pod, curScore int, tolerance int, nodes []*v1.Node) (*v1.Node, int, error) {
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
