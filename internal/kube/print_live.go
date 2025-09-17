// kube/print_live.go
package kube

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/olekukonko/tablewriter"
)

type LiveTablePrinter struct {
	out      io.Writer
	mu       sync.Mutex
	lines    int
	rendered bool
}

func NewLiveTablePrinter(writer io.Writer) *LiveTablePrinter {
	return &LiveTablePrinter{out: writer}
}

func (t *LiveTablePrinter) Print(rows []PodRow) error   { return t.render(rows, false) }
func (t *LiveTablePrinter) Refresh(rows []PodRow) error { return t.render(rows, true) }

func (t *LiveTablePrinter) render(rows []PodRow, inplace bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Render the table to a buffer
	buf := &bytes.Buffer{}
	tw := tablewriter.NewWriter(buf)
	tw.Header([]string{"NAME", "NAMESPACE", "READY", "STATUS", "RESTARTS", "AGE", "NODE"})

	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		data = append(data, []string{r.Name, r.Namespace, r.Ready, r.Status, r.Restarts, r.Age, r.Node})
	}
	tw.Bulk(data)
	tw.Render()

	frame := buf.String()
	// Count lines to know how far to move the cursor up next time.
	// Trim trailing newline so the count matches visible lines.
	trimmed := strings.TrimRight(frame, "\n")
	lines := strings.Count(trimmed, "\n") + 1

	// Hide cursor during repaint to reduce flicker
	fmt.Fprint(t.out, "\x1b[?25l")
	defer fmt.Fprint(t.out, "\x1b[?25h")

	if inplace && t.rendered && t.lines > 0 {
		// Move cursor up to the start of previous frame and clear to end of screen
		// \x1b[{n}A = move up n lines, \r = CR to column 0, \x1b[J = clear to end of screen
		fmt.Fprintf(t.out, "\x1b[%dA\r\x1b[J", t.lines)
	}

	fmt.Fprint(t.out, frame)

	t.lines = lines
	t.rendered = true
	return nil
}
