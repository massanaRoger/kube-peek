package kube

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestCalcAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		creationTime time.Time
		expected     string
	}{
		{
			name:         "30 seconds ago",
			creationTime: now.Add(-30 * time.Second),
			expected:     "30s",
		},
		{
			name:         "1 second ago",
			creationTime: now.Add(-1 * time.Second),
			expected:     "1s",
		},
		{
			name:         "59 seconds ago",
			creationTime: now.Add(-59 * time.Second),
			expected:     "59s",
		},
		{
			name:         "1 minute ago",
			creationTime: now.Add(-1 * time.Minute),
			expected:     "1m",
		},
		{
			name:         "30 minutes ago",
			creationTime: now.Add(-30 * time.Minute),
			expected:     "30m",
		},
		{
			name:         "59 minutes ago",
			creationTime: now.Add(-59 * time.Minute),
			expected:     "59m",
		},
		{
			name:         "1 hour ago",
			creationTime: now.Add(-1 * time.Hour),
			expected:     "1h",
		},
		{
			name:         "12 hours ago",
			creationTime: now.Add(-12 * time.Hour),
			expected:     "12h",
		},
		{
			name:         "23 hours ago",
			creationTime: now.Add(-23 * time.Hour),
			expected:     "23h",
		},
		{
			name:         "1 day ago",
			creationTime: now.Add(-24 * time.Hour),
			expected:     "1d",
		},
		{
			name:         "5 days ago",
			creationTime: now.Add(-5 * 24 * time.Hour),
			expected:     "5d",
		},
		{
			name:         "30 days ago",
			creationTime: now.Add(-30 * 24 * time.Hour),
			expected:     "30d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcAge(tt.creationTime)
			if result != tt.expected {
				t.Errorf("calcAge(%v) = %q, want %q",
					tt.creationTime.Format("2006-01-02 15:04:05"),
					result,
					tt.expected)
			}
		})
	}
}

// Mock data

type mockPodSource struct {
	listResult  *v1.PodList
	listError   error
	watchResult watch.Interface
	watchError  error
}

func (m *mockPodSource) List(ctx context.Context, ns string, opts ListOpts) (*v1.PodList, error) {
	if m.listError != nil {
		return nil, m.listError
	}
	return m.listResult, nil
}

func (m *mockPodSource) Watch(ctx context.Context, ns string, opts ListOpts) (watch.Interface, error) {
	if m.watchError != nil {
		return nil, m.watchError
	}
	return m.watchResult, nil
}

type mockPrinter struct {
	printCalled   bool
	refreshCalled bool
	printError    error
	refreshError  error
	lastRows      []PodRow
}

func (m *mockPrinter) Print(rows []PodRow) error {
	m.printCalled = true
	m.lastRows = rows
	return m.printError
}

func (m *mockPrinter) Refresh(rows []PodRow) error {
	m.refreshCalled = true
	m.lastRows = rows
	return m.refreshError
}

type mockWatcher struct {
	resultChan chan watch.Event
	stopped    bool
}

func (m *mockWatcher) Stop() {
	m.stopped = true
	close(m.resultChan)
}

func (m *mockWatcher) ResultChan() <-chan watch.Event {
	return m.resultChan
}

func TestController_Run(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func() (*mockPodSource, *mockPrinter)
		runOpts       RunOpts
		expectError   bool
		expectPrint   bool
		expectRefresh bool
	}{
		{
			name: "successful list only",
			setupMocks: func() (*mockPodSource, *mockPrinter) {
				source := &mockPodSource{
					listResult: &v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:              "test-pod",
									Namespace:         "default",
									CreationTimestamp: metav1.Time{Time: time.Now()},
								},
								Status: v1.PodStatus{
									Phase: v1.PodRunning,
									ContainerStatuses: []v1.ContainerStatus{
										{Ready: true, RestartCount: 0},
									},
								},
							},
						},
					},
				}
				printer := &mockPrinter{}
				return source, printer
			},
			runOpts: RunOpts{
				Namespace: "default",
				Watch:     false,
			},
			expectError:   false,
			expectPrint:   true,
			expectRefresh: false,
		},
		{
			name: "list error",
			setupMocks: func() (*mockPodSource, *mockPrinter) {
				source := &mockPodSource{
					listError: errors.New("failed to list pods"),
				}
				printer := &mockPrinter{}
				return source, printer
			},
			runOpts: RunOpts{
				Namespace: "default",
				Watch:     false,
			},
			expectError:   true,
			expectPrint:   false,
			expectRefresh: false,
		},
		{
			name: "print error",
			setupMocks: func() (*mockPodSource, *mockPrinter) {
				source := &mockPodSource{
					listResult: &v1.PodList{Items: []v1.Pod{}},
				}
				printer := &mockPrinter{
					printError: errors.New("failed to print"),
				}
				return source, printer
			},
			runOpts: RunOpts{
				Namespace: "default",
				Watch:     false,
			},
			expectError:   true,
			expectPrint:   true,
			expectRefresh: false,
		},
		{
			name: "watch mode with cancelled context",
			setupMocks: func() (*mockPodSource, *mockPrinter) {
				watcher := &mockWatcher{
					resultChan: make(chan watch.Event),
				}
				source := &mockPodSource{
					listResult:  &v1.PodList{Items: []v1.Pod{}},
					watchResult: watcher,
				}
				printer := &mockPrinter{}
				return source, printer
			},
			runOpts: RunOpts{
				Namespace: "default",
				Watch:     true,
			},
			expectError:   false,
			expectPrint:   true,
			expectRefresh: false,
		},
		{
			name: "watch error",
			setupMocks: func() (*mockPodSource, *mockPrinter) {
				source := &mockPodSource{
					listResult: &v1.PodList{Items: []v1.Pod{}},
					watchError: errors.New("failed to watch"),
				}
				printer := &mockPrinter{}
				return source, printer
			},
			runOpts: RunOpts{
				Namespace: "default",
				Watch:     true,
			},
			expectError:   true,
			expectPrint:   true,
			expectRefresh: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, printer := tt.setupMocks()

			controller := Controller{
				Source:         source,
				CurrentPrinter: printer,
			}

			// Create cancellable context for watch
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// For watch mode tests, cancel context quickly to avoid hanging
			if tt.runOpts.Watch && !tt.expectError {
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			err := controller.Run(ctx, tt.runOpts)

			if tt.expectError && err == nil {
				t.Errorf("Controller.Run() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Controller.Run() unexpected error: %v", err)
			}

			if tt.expectPrint && !printer.printCalled {
				t.Errorf("Controller.Run() expected Print() to be called")
			}
			if !tt.expectPrint && printer.printCalled {
				t.Errorf("Controller.Run() unexpected Print() call")
			}

			if tt.expectRefresh && !printer.refreshCalled {
				t.Errorf("Controller.Run() expected Refresh() to be called")
			}
			if !tt.expectRefresh && printer.refreshCalled {
				t.Errorf("Controller.Run() unexpected Refresh() call")
			}
		})
	}
}
