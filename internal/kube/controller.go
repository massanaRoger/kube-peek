package kube

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
}
