package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func ListPods(client kubernetes.Interface) ([]*v1.Pod, error) {
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
		} else {
			klog.V(4).InfoS("Skipping pod without PreferredDuringSchedulingIgnoredDuringExecution terms", "pod", klog.KObj(pod))
		}
	}
	return pods, nil
}

func ListPodsOnNode(client kubernetes.Interface, nodeName string) ([]*v1.Pod, error) {
	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed) + ",status.phase!=" + string(v1.PodPending) + ",spec.nodeName=" + nodeName)

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
		pods = append(pods, pod)
	}
	return pods, nil
}
