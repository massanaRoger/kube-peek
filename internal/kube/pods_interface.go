package kube

import (
	"context"
	"io"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type PodSource interface {
	List(ctx context.Context, ns string, opts ListOpts) (*v1.PodList, error)
	Watch(ctx context.Context, ns string, opts ListOpts) (watch.Interface, error)
}

type ListOpts struct {
	LabelSelector   string
	FieldSelector   string
	ResourceVersion string
}

type PodRow struct {
	Name      string
	Namespace string
	Ready     string
	Status    string
	Restarts  string
	Age       string
	Node      string
}
