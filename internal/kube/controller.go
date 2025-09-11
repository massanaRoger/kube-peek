package kube

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	v1 "k8s.io/api/core/v1"
)

type Controller struct {
	Source  PodSource
	Printer Printer
}

type RunOpts struct {
	Namespace string
	ListOpts  ListOpts
	Watch     bool
}

func (c Controller) Run(ctx context.Context, opts RunOpts) error {
	// 1) Initial LIST
	list, err := c.Source.List(ctx, opts.Namespace, opts.ListOpts)
	if err != nil {
		return err
	}

	// Snapshot -> rows -> print
	rows := ToRows(list.Items)
	if err := c.Printer.Print(rows); err != nil {
		return err
	}
	if !opts.Watch {
		return nil
	}

	// 2) WATCH from the same ResourceVersion
	w, err := c.Source.(ClientGoSource).Watch(ctx, opts.Namespace, opts.ListOpts)

	if err != nil {
		return err
	}
	defer w.Stop()

	// Keep a local store keyed by ns/name
	store := map[string]v1.Pod{}
	for _, p := range list.Items {
		store[p.Namespace+"/"+p.Name] = p
	}

	// Cancel on Ctrl+C too
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-w.ResultChan():
			if !ok { return nil } // stream closed

			switch obj := ev.Object.(type) {
			case *v1.Pod:
				key := obj.Namespace + "/" + obj.Name
				switch ev.Type {
				case "ADDED", "MODIFIED":
					store[key] = *obj
				case "DELETED":
					delete(store, key)
				}
				// Re-render snapshot
				snap := make([]v1.Pod, 0, len(store))
				for _, p := range store { snap = append(snap, p) }
				if err := c.Printer.Refresh(ToRows(snap)); err != nil { return err }
			}
		}
	}
}
