package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestYoungestPodAge(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		api         KubernetesAPI
		expectedAge ResourceAge
		found       bool
		wantErr     bool
	}{
		{
			name: "get correct pod age 10h",
			api: &KubernetesAPIMock{
				pods: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: now.Add(-10 * time.Hour),
							},
						},
					},
				},
			},
			expectedAge: ResourceAge(36000),
			found:       true,
			wantErr:     false,
		},
		{
			name: "get correct pod age 10h",
			api: &KubernetesAPIMock{
				pods: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: now.Add(-5 * time.Hour),
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: now.Add(-10 * time.Hour),
							},
						},
					},
				},
			},
			expectedAge: 18000,
			found:       true,
			wantErr:     false,
		},
		{
			name: "empty pod list - expect error",
			api: &KubernetesAPIMock{
				pods: []v1.Pod{},
			},
			expectedAge: 0,
			found:       false,
			wantErr:     false,
		},
		{
			name: "expect list error (from k8s client side)",
			api: &KubernetesAPIMock{
				err: errors.New("pseudo random k8s appsv1 pods list error"),
			},
			expectedAge: 0,
			found:       false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := YoungestPodAge(context.TODO(), tt.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoungestPodAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedAge {
				t.Errorf("YoungestPodAge() got = %v, want %v", got, tt.expectedAge)
			}
			if got1 != tt.found {
				t.Errorf("YoungestPodAge() got1 = %v, want %v", got1, tt.found)
			}
		})
	}
}
