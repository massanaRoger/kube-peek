package kube

import (
	"context"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListPods(client *kubernetes.Clientset, table *tablewriter.Table, namespace string) error {
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
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
