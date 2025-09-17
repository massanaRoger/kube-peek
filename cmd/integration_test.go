package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/massanaRoger/kube-peek/internal/kube"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)


func TestEndToEndPodListing(t *testing.T) {
	now := time.Now()
	
	mockPods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-deployment-12345",
				Namespace: "default",
				CreationTimestamp: metav1.Time{Time: now.Add(-2 * time.Hour)},
			},
			Spec: v1.PodSpec{
				NodeName: "worker-1",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{Ready: true, RestartCount: 0},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis-cache",
				Namespace: "default",
				CreationTimestamp: metav1.Time{Time: now.Add(-30 * time.Minute)},
			},
			Spec: v1.PodSpec{
				NodeName: "worker-2",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{Ready: true, RestartCount: 1},
					{Ready: false, RestartCount: 0},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-proxy-xyz",
				Namespace: "kube-system",
				CreationTimestamp: metav1.Time{Time: now.Add(-24 * time.Hour)},
			},
			Spec: v1.PodSpec{
				NodeName: "master",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{Ready: true, RestartCount: 0},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pending-pod",
				Namespace: "default",
				CreationTimestamp: metav1.Time{Time: now.Add(-5 * time.Minute)},
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
				ContainerStatuses: []v1.ContainerStatus{},
			},
		},
	}

	tests := []struct {
		name           string
		namespace      string
		allNamespaces  bool
		outputFormat   string
		expectedInOutput []string
		notInOutput    []string
	}{
		{
			name:         "default namespace only",
			namespace:    "default",
			allNamespaces: false,
			outputFormat: "table",
			expectedInOutput: []string{
				"nginx-deployment-12345",
				"redis-cache", 
				"pending-pod",
				"worker-1",
				"worker-2",
				"Running",
				"Pending",
				"1/2", // redis has 1 ready out of 2 containers
				"0/0", // pending pod has no containers
			},
			notInOutput: []string{
				"kube-proxy-xyz", // should not appear (different namespace)
				"kube-system",
			},
		},
		{
			name:         "kube-system namespace",
			namespace:    "kube-system",
			allNamespaces: false,
			outputFormat: "table",
			expectedInOutput: []string{
				"kube-proxy-xyz",
				"kube-system",
				"master",
			},
			notInOutput: []string{
				"nginx-deployment-12345",
				"redis-cache",
				"default",
			},
		},
		{
			name:         "all namespaces",
			namespace:    "",
			allNamespaces: true,
			outputFormat: "table",
			expectedInOutput: []string{
				"nginx-deployment-12345",
				"redis-cache",
				"kube-proxy-xyz",
				"pending-pod",
				"default",
				"kube-system",
				"worker-1",
				"worker-2",
				"master",
			},
			notInOutput: []string{},
		},
		{
			name:         "json output",
			namespace:    "default",
			allNamespaces: false,
			outputFormat: "json",
			expectedInOutput: []string{
				`"Name": "nginx-deployment-12345"`,
				`"Namespace": "default"`,
				`"Status": "Running"`,
				`"Node": "worker-1"`,
			},
			notInOutput: []string{
				"kube-proxy-xyz",
			},
		},
		{
			name:         "empty namespace",
			namespace:    "nonexistent",
			allNamespaces: false,
			outputFormat: "table",
			expectedInOutput: []string{
				"NAME", "NAMESPACE", // headers should still appear
			},
			notInOutput: []string{
				"nginx-deployment-12345",
				"redis-cache",
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
					t.Fatalf("Failed to create fake pod: %v", err)
				}
			}
			
			source := kube.ClientGoSource{Client: fakeClient}
			
			var printer kube.Printer
			if tt.outputFormat == "json" {
				printer = kube.NewJsonPrinter(&buf)
			} else {
				printer = kube.NewTablePrinter(&buf)
			}

			controller := kube.Controller{
				Source:         source,
				CurrentPrinter: printer,
			}

			err := controller.Run(context.Background(), kube.RunOpts{
				Namespace: tt.namespace,
				ListOpts:  kube.ListOpts{},
				Watch:     false,
			})

			if err != nil {
				t.Errorf("Controller.Run() error = %v", err)
				return
			}

			output := buf.String()

			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q\nActual output:\n%s", expected, output)
				}
			}

			for _, notExpected := range tt.notInOutput {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output to NOT contain %q\nActual output:\n%s", notExpected, output)
				}
			}
		})
	}
}

func TestEndToEndErrorHandling(t *testing.T) {
	t.Run("printer error", func(t *testing.T) {
		fakeClient := fake.NewSimpleClientset()
		source := kube.ClientGoSource{Client: fakeClient}
		
		controller := kube.Controller{
			Source:         source,
			CurrentPrinter: &mockErrorPrinter{},
		}

		err := controller.Run(context.Background(), kube.RunOpts{
			Namespace: "default",
			ListOpts:  kube.ListOpts{},
			Watch:     false,
		})

		if err == nil {
			t.Error("Expected error but got none")
		}
	})
}

type mockErrorPrinter struct{}

func (m *mockErrorPrinter) Print(rows []kube.PodRow) error {
	return context.DeadlineExceeded
}

func (m *mockErrorPrinter) Refresh(rows []kube.PodRow) error {
	return context.DeadlineExceeded
}