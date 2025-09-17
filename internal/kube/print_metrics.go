package kube

import (
	"encoding/json"
	"io"

	"github.com/olekukonko/tablewriter"
)

type MetricsPrinter struct {
	Writer io.Writer
}

func NewMetricsPrinter(writer io.Writer) MetricsPrinter {
	return MetricsPrinter{Writer: writer}
}

func (p MetricsPrinter) Print(rows []PodMetricsRow) error {
	return p.render(rows)
}

func (p MetricsPrinter) Refresh(rows []PodMetricsRow) error {
	return p.render(rows)
}

func (p MetricsPrinter) render(rows []PodMetricsRow) error {
	table := tablewriter.NewWriter(p.Writer)
	table.Header([]string{"NAME", "CPU", "MEMORY"})
	
	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		data = append(data, []string{r.Name, r.CPU, r.Memory})
	}
	table.Bulk(data)
	table.Render()
	return nil
}

type MetricsJSONPrinter struct {
	Writer io.Writer
}

func NewMetricsJSONPrinter(writer io.Writer) MetricsJSONPrinter {
	return MetricsJSONPrinter{Writer: writer}
}

func (p MetricsJSONPrinter) Print(rows []PodMetricsRow) error {
	enc := json.NewEncoder(p.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(rows)
}

func (p MetricsJSONPrinter) Refresh(rows []PodMetricsRow) error {
	return p.Print(rows)
}