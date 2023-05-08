package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

func main() {
	var (
		dryRun    = flag.Bool("dry-run", false, "Dry Run")
		tolerance = flag.Int("tolerance", 0, "Ignore certain weight difference")
	)
	flag.Parse()

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		logrus.Fatalf("Couldn't get Kubernetes default config: %s", err)
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	hasPotential := false

	nodes, err := ListNodes(client)
	if err != nil {
		panic(err.Error())
	}

	pods, err := ListPods(client)
	if err != nil {
		panic(err.Error())
	}

	for _, pod := range pods {
		node, err := client.CoreV1().Nodes().Get(context.Background(), pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}

		curScore, err := CalculatePodPriorityScore(pod, node)
		if err != nil {
			panic(err.Error())
		}

		foundBetter, _, nodeNameBetter := FindBetterNode(pod, curScore, *tolerance, nodes)
		if foundBetter {
			hasPotential = true
			logrus.Infof("Pod %v/%v can possibly be scheduled on %v", pod.Namespace, pod.Name, nodeNameBetter)
			if !*dryRun {
				err := client.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					panic(err.Error())
				}
				logrus.Infof("Pod %v/%v has been evicted!", pod.Namespace, pod.Name)
			}
		}
	}

	if !hasPotential {
		logrus.Info("No Pods to evict")
	}
}

func ListNodes(client clientset.Interface) ([]*v1.Node, error) {
	nodeList, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []*v1.Node{}, err
	}

	nodes := make([]*v1.Node, 0)
	for i := range nodeList.Items {
		node := &nodeList.Items[i]

		if node.Spec.Unschedulable {
			logrus.Infof("Not evaluating unschedulable node %v", node.Name)
			continue
		}

		nodes = append(nodes, node)
	}
	return nodes, nil
}

func ListPods(client clientset.Interface) ([]*v1.Pod, error) {
	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed) + ",status.phase!=" + string(v1.PodPending))
	if err != nil {
		return []*v1.Pod{}, err
	}

	podList, err := client.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return []*v1.Pod{}, err
	}

	pods := make([]*v1.Pod, 0)
	for i := range podList.Items {
		pod := &podList.Items[i]
		affinity := pod.Spec.Affinity

		if affinity != nil && affinity.NodeAffinity != nil && affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
			pods = append(pods, pod)
		}
	}
	return pods, nil
}

func CalculatePodPriorityScore(pod *v1.Pod, node *v1.Node) (int, error) {
	var score int32
	affinity := pod.Spec.Affinity

	for _, preferredSchedulingTerm := range affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		if preferredSchedulingTerm.Weight == 0 {
			continue
		}

		selector, err := v1helper.NodeSelectorRequirementsAsSelector(preferredSchedulingTerm.Preference.MatchExpressions)
		if err != nil {
			return 0, err
		}

		if selector.Matches(labels.Set(node.Labels)) {
			score += preferredSchedulingTerm.Weight
		}
	}

	return int(score), nil
}

func FindBetterNode(pod *v1.Pod, curScore int, tolerance int, nodes []*v1.Node) (bool, int, string) {
	for _, node := range nodes {

		// Skip nodes that do not match the Pod's NodeSelector
		if len(pod.Spec.NodeSelector) > 0 {
			nodeSelector := labels.SelectorFromSet(pod.Spec.NodeSelector)
			if !nodeSelector.Matches(labels.Set(node.Labels)) {
				continue
			}
		}

		// Skip nodes that do not match the Pod's required nodeAffinity
		nodeAffinity := pod.Spec.Affinity.NodeAffinity
		if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			if !v1helper.MatchNodeSelectorTerms(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, labels.Set(node.Labels), nil) {
				continue
			}
		}

		score, err := CalculatePodPriorityScore(pod, node)
		if err != nil {
			continue
		}

		if (score - tolerance) > curScore {
			return true, score, node.Name
		}
	}

	return false, 0, ""
}
