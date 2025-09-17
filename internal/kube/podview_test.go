package kube

import (
	"testing"
	"time"
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
