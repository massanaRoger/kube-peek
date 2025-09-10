package kube

import (
	"context"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ListPodsFlags struct {
	Selector      string
	FieldSelector string
	Watch         bool
}

func ListPods(client *kubernetes.Clientset, table *tablewriter.Table, namespace string, flags ListPodsFlags) error {
	if flags.Watch {
		watchCache := NewCache()
		err := watchCache.WatchPods(client, table, namespace, flags)
		return err
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: flags.Selector, FieldSelector: flags.FieldSelector, Watch: flags.Watch})

	if err != nil {
		return err
	}
	tableData := make([][]string, pods.Size())
	for _, p := range pods.Items {
		row := []string{p.Name, p.Namespace}
		tableData = append(tableData, row)
	}

	table.Header([]string{"NAME", "NAMESPACE"})
	table.Bulk(tableData)
	table.Render()
	return nil
}
