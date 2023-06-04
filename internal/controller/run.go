package controller

import (
	"context"

	"github.com/decayofmind/kube-better-node/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
	"k8s.io/klog/v2"
)

func Run(client kubernetes.Interface, dryRun bool, tolerance int) {
	hasPotential := false

	nodes, err := k8s.ListNodes(client)
	if err != nil {
		panic(err.Error())
	}

	pods, err := k8s.ListPods(client)
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods {
		node, err := client.CoreV1().Nodes().Get(context.Background(), pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		nodeAffinity := pod.Spec.Affinity.NodeAffinity

		terms, err := nodeaffinity.NewPreferredSchedulingTerms(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
		if err != nil {
			panic(err.Error())
		}

		curScore := int(terms.Score(node))

		if err != nil {
			panic(err.Error())
		}

		betterNode, _, err := FindBetterNode(pod, curScore, tolerance, nodes)
		if err != nil {
			panic(err.Error())
		}

		if betterNode != nil {
			hasPotential = true
			klog.InfoS("Found better node candidate for pod", "pod", klog.KObj(pod), "node", klog.KObj(betterNode))
			if !dryRun {
				err := client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					panic(err.Error())
				}
				klog.InfoS("Evicted pod", "pod", klog.KObj(pod))
			} else {
				klog.InfoS("Evicted pod (dry run)", "pod", klog.KObj(pod))
			}
		}
	}

	if !hasPotential {
		klog.InfoS("No Pods to evict")
	}

}
