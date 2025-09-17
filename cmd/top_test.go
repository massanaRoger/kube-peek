package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/massanaRoger/kube-peek/internal/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestTopPodsCommand(t *testing.T) {
	mockPods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-pod",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "nginx"},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-pod", 
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "redis"},
				},
			},
		},
	}

	mockMetrics := []metricsv1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-pod",
				Namespace: "default",
			},
			Containers: []metricsv1beta1.ContainerMetrics{
				{
					Name: "nginx",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    parseQuantity("100m"),
						v1.ResourceMemory: parseQuantity("128Mi"),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-pod",
				Namespace: "default",
			},
			Containers: []metricsv1beta1.ContainerMetrics{
				{
					Name: "redis",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    parseQuantity("50m"),
						v1.ResourceMemory: parseQuantity("64Mi"),
					},
				},
			},
		},
	}

	tests := []struct {
		name         string
		namespace    string
		expectedIn   []string
		notExpected  []string
	}{
		{
			name:      "default namespace pods",
			namespace: "default",
			expectedIn: []string{
				"nginx-pod", "redis-pod",
				"100m", "50m",
				"128Mi", "64Mi",
				"CPU", "MEMORY", // headers
			},
			notExpected: []string{},
		},
		{
			name:      "empty namespace",
			namespace: "nonexistent",
			expectedIn: []string{
				"NAME", "CPU", "MEMORY", // headers should appear
			},
			notExpected: []string{
				"nginx-pod", "redis-pod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			fakeClient := fake.NewSimpleClientset()
			
			for _, pod := range mockPods {
				_, err := fakeClient.CoreV1().Pods(pod.Namespace).Create(
					context.Background(), &pod, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create pod: %v", err)
				}
			}

			printer := kube.NewMetricsPrinter(&buf)

			controller := &TestMetricsController{
				client:  fakeClient,
				metrics: mockMetrics,
				printer: printer,
			}

			err := controller.Run(context.Background(), tt.namespace)

			if err != nil {
				t.Errorf("MetricsController.Run() error = %v", err)
				return
			}

			output := buf.String()

			for _, expected := range tt.expectedIn {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q\nActual output:\n%s", expected, output)
				}
			}

			for _, notExpected := range tt.notExpected {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output to NOT contain %q\nActual output:\n%s", notExpected, output)
				}
			}
		})
	}
}

func TestTopPodsErrorHandling(t *testing.T) {
	t.Run("no metrics server", func(t *testing.T) {
		var buf bytes.Buffer

		fakeClient := fake.NewSimpleClientset()
		printer := kube.NewMetricsPrinter(&buf)

		controller := &TestMetricsController{
			client:  fakeClient,
			metrics: []metricsv1beta1.PodMetrics{},
			printer: printer,
		}

		err := controller.Run(context.Background(), "default")

		if err != nil {
			t.Errorf("Should handle empty metrics gracefully, got error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "NAME") {
			t.Errorf("Should show headers even with no data")
		}
	})
}

type TestMetricsController struct {
	client  *fake.Clientset
	metrics []metricsv1beta1.PodMetrics
	printer kube.MetricsPrinter
}

func (c *TestMetricsController) Run(ctx context.Context, namespace string) error {
	podList, err := c.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	metricsMap := make(map[string]metricsv1beta1.PodMetrics)
	for _, metrics := range c.metrics {
		if namespace == "" || metrics.Namespace == namespace {
			metricsMap[metrics.Name] = metrics
		}
	}

	var rows []kube.PodMetricsRow
	for _, pod := range podList.Items {
		metrics, hasMetrics := metricsMap[pod.Name]
		
		var cpu, memory string
		if hasMetrics {
			cpu, memory = kube.CalculatePodUsage(metrics)
		} else {
			cpu, memory = "<unknown>", "<unknown>"
		}

		rows = append(rows, kube.PodMetricsRow{
			Name:   pod.Name,
			CPU:    cpu,
			Memory: memory,
		})
	}

	return c.printer.Print(rows)
}


func parseQuantity(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q
}