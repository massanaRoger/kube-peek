package kube

import (
	"context"
	"fmt"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type InterfaceCache interface {
	DisplayWatchPods(table *tablewriter.Table)
	WatchPods(client *kubernetes.Clientset, table *tablewriter.Table, namespace string, flags ListPodsFlags) error
}

type Cache struct {
	Pods map[string]*v1.Pod
}

func NewCache() Cache {
	return Cache{
		Pods: make(map[string]*v1.Pod),
	}
}

func (c *Cache) DisplayWatchPods(table *tablewriter.Table) {
	fmt.Print("\033[2J\033[H")

	table.Reset()

	tableData := make([][]string, len(c.Pods))
	for _, p := range c.Pods {
		row := []string{p.Name, p.Namespace}
		tableData = append(tableData, row)
	}

	table.Header([]string{"NAME", "NAMESPACE"})
	table.Bulk(tableData)
	table.Render()
}

func (c *Cache) WatchPods(client *kubernetes.Clientset, table *tablewriter.Table, namespace string, flags ListPodsFlags) error {

	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: flags.Selector, FieldSelector: flags.FieldSelector})
	if err != nil {
		return err
	}

	watchEvent, err := client.CoreV1().Pods(namespace).Watch(context.TODO(), metav1.ListOptions{LabelSelector: flags.Selector, FieldSelector: flags.FieldSelector, ResourceVersion: pods.ResourceVersion})
	for val := range watchEvent.ResultChan() {
		pod, ok := val.Object.(*v1.Pod)
		if !ok {
			fmt.Println("Unexpected type")
			continue
		}

		switch val.Type {
		case watch.Added:
			c.Pods[pod.Name] = pod
		case watch.Modified:
		case watch.Deleted:
			delete(c.Pods, pod.Name)
		}

		c.DisplayWatchPods(table)
	}

	return nil
}
