package kube

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type ClientGoSource struct{ Client *kubernetes.Clientset }

func (s ClientGoSource) List(ctx context.Context, ns string, opts ListOpts) (*v1.PodList, error) {
	return s.Client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: opts.LabelSelector,
		FieldSelector: opts.FieldSelector,
	})
}

func (s ClientGoSource) Watch(ctx context.Context, ns string, opts ListOpts) (watch.Interface, error) {
	return s.Client.CoreV1().Pods(ns).Watch(ctx, metav1.ListOptions{
		LabelSelector:       opts.LabelSelector,
		FieldSelector:       opts.FieldSelector,
		ResourceVersion:     opts.ResourceVersion,
		AllowWatchBookmarks: true,
		// NOTE: pass ResourceVersion from the initial List for consistent watch (see controller below)
	})
}
