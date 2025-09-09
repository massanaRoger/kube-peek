package kube

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
)

type Output struct {
	Table *tablewriter.Table
	Type  string
}

func (o *Output) InitWriter() {
	o.Table = tablewriter.NewWriter(os.Stdout)
}

func (o *Output) PrintPods(pods *v1.PodList) error {
	switch o.Type {
	case "", "table":
		o.renderTable(pods)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pods) // print the raw API object
	default:
		return fmt.Errorf("invalid --output %q (use: table|json)", o.Type)
	}
	return nil
}

func (o *Output) renderTable(pods *v1.PodList) {
	tableData := make([][]string, pods.Size())
	for _, p := range pods.Items {
		row := []string{p.Name, p.Namespace, readiness(p.Status.ContainerStatuses), string(p.Status.Phase), containerRestarts(p.Status.ContainerStatuses), calcAge(p.CreationTimestamp.Time), p.Spec.NodeName}
		tableData = append(tableData, row)
	}

	o.Table.Header([]string{"NAME", "NAMESPACE", "READY", "STATUS", "RESTARTS", "AGE", "NODE"})
	o.Table.Bulk(tableData)
	o.Table.Render()
}

func (o *Output) RenderJSON(pods *v1.PodList) {

}
