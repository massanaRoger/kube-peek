package kube

import (
	"encoding/json"
	"io"

	"github.com/olekukonko/tablewriter"
)

type Printer interface {
	Print([]PodRow) error   // print the current snapshot
	Refresh([]PodRow) error // re-print (watch mode); can be same as Print
}

type TablePrinter struct {
	Writer io.Writer
}
type JSONPrinter struct {
	Writer io.Writer
}

func NewTablePrinter(writer io.Writer) TablePrinter {
	return TablePrinter{
		Writer: writer,
	}
}

func NewJsonPrinter(writer io.Writer) JSONPrinter {
	return JSONPrinter{
		Writer: writer,
	}
}

func (p TablePrinter) Print(rows []PodRow) error   { return p.render(rows) }
func (p TablePrinter) Refresh(rows []PodRow) error { return p.render(rows) }

func (p TablePrinter) render(rows []PodRow) error {
	table := tablewriter.NewWriter(p.Writer)
	table.Header([]string{"NAME", "NAMESPACE", "READY", "STATUS", "RESTARTS", "AGE", "NODE"})
	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		data = append(data, []string{r.Name, r.Namespace, r.Ready, r.Status, r.Restarts, r.Age, r.Node})
	}
	table.Bulk(data)
	table.Render()
	return nil
}

func (p JSONPrinter) Print(rows []PodRow) error {
	enc := json.NewEncoder(p.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}
func (p JSONPrinter) Refresh(rows []PodRow) error { return p.Print(rows) }
