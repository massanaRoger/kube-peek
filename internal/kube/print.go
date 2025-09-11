package kube

import (
	"encoding/json"
	"os"

	"github.com/olekukonko/tablewriter"
)

type TablePrinter struct{}
type JSONPrinter struct{}

func (t TablePrinter) Print(rows []PodRow) error   { return t.render(rows) }
func (t TablePrinter) Refresh(rows []PodRow) error { return t.render(rows) }

func (t TablePrinter) render(rows []PodRow) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NAME", "NAMESPACE", "READY", "STATUS", "RESTARTS", "AGE", "NODE"})
	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		data = append(data, []string{r.Name, r.Namespace, r.Ready, r.Status, r.Restarts, r.Age, r.Node})
	}
	table.Bulk(data)
	table.Render()
	return nil
}

func (JSONPrinter) Print(rows []PodRow) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
func (JSONPrinter) Refresh(rows []PodRow) error { return (JSONPrinter{}).Print(rows) }
