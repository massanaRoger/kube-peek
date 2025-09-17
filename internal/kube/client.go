package kube

import (
	"flag"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Provider interface {
	ClientSet() (kubernetes.Interface, error)
	MetricsClient() (metricsclientset.Interface, error)
}

type provider struct {
	client        kubernetes.Interface
	metricsClient metricsclientset.Interface
}

func (f *provider) ClientSet() (kubernetes.Interface, error) {
	return f.client, nil
}

func (f *provider) MetricsClient() (metricsclientset.Interface, error) {
	return f.metricsClient, nil
}

func NewProvider() (*provider, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClientset, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &provider{
		client:        clientset,
		metricsClient: metricsClientset,
	}, nil
}

