package kube

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
)

func ToRows(pods []v1.Pod) []PodRow {
	rows := make([]PodRow, 0, len(pods))
	for _, p := range pods {
		rows = append(rows, PodRow{
			Name:      p.Name,
			Namespace: p.Namespace,
			Ready:     readiness(p.Status.ContainerStatuses),
			Status:    string(p.Status.Phase),
			Restarts:  containerRestarts(p.Status.ContainerStatuses),
			Age:       calcAge(p.CreationTimestamp.Time),
			Node:      p.Spec.NodeName,
		})
	}
	return rows
}

func readiness(sts []v1.ContainerStatus) string {
	ready := 0
	for _, s := range sts {
		if s.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, len(sts))
}

func containerRestarts(sts []v1.ContainerStatus) string {
	r := 0
	for _, s := range sts {
		r += int(s.RestartCount)
	}
	return fmt.Sprintf("%d", r)
}

func calcAge(creation time.Time) string {
	d := time.Since(creation.UTC())
	sec := int64(d.Seconds())
	min := sec / 60
	hr := min / 60
	day := hr / 24
	switch {
	case sec < 60:
		return fmt.Sprintf("%ds", sec)
	case min < 60:
		return fmt.Sprintf("%dm", min)
	case hr < 24:
		return fmt.Sprintf("%dh", hr)
	default:
		return fmt.Sprintf("%dd", day)
	}
}
