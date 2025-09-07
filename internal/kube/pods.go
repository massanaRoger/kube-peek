package kube

import (
	"context"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ListPodsFlags struct {
	Selector      string
	FieldSelector string
	Watch         bool
}

func ListPods(client *kubernetes.Clientset, table *tablewriter.Table, namespace string, flags ListPodsFlags) error {
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: flags.Selector, FieldSelector: flags.FieldSelector, Watch: flags.Watch})
	if err != nil {
		return err
	}
	tableData := make([][]string, pods.Size())
	for _, p := range pods.Items {
		row := []string{p.Name, p.Namespace, readiness(p.Status.ContainerStatuses), string(p.Status.Phase), containerRestarts(p.Status.ContainerStatuses), calcAge(p.CreationTimestamp.Time), p.Spec.NodeName}
		tableData = append(tableData, row)
	}

	table.Header([]string{"NAME", "NAMESPACE", "READY", "STATUS", "RESTARTS", "AGE", "NODE"})
	table.Bulk(tableData)
	table.Render()
	return nil
}

func readiness(statuses []v1.ContainerStatus) string {
	readyCount := 0
	for _, status := range statuses {
		if status.Ready {
			readyCount += 1
		}
	}

	return fmt.Sprintf("%d/%d", readyCount, len(statuses))
}

func containerRestarts(statuses []v1.ContainerStatus) string {
	restarts := 0
	for _, status := range statuses {
		restarts += int(status.RestartCount)
	}

	return fmt.Sprintf("%d", restarts)
}

func calcAge(creation time.Time) string {
	now := time.Now().UTC()
	delta := now.Sub(creation)
	return humanDuration(delta)
}

func humanDuration(d time.Duration) string {
	seconds := int64(d.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case minutes < 60:
		return fmt.Sprintf("%dm", minutes)
	case hours < 24:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dd", days)
	}
}
