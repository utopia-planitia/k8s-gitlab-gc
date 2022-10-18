package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestYoungestDaemonsetAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		api         KubernetesAPI
		expectedAge ResourceAge
		found       bool
		wantErr     bool
	}{
		{
			name: "expect list error (from k8s client side)",
			api: &KubernetesAPIMock{
				err: errors.New("pseudo random k8s appsv1 daemonsets list error"),
			},
			expectedAge: 0,
			found:       false,
			wantErr:     true,
		},
		{
			name: "get correct daemonset age 10h",
			api: &KubernetesAPIMock{
				daemonSet: []appsv1.DaemonSet{
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
			name: "get correct daemonset age 5h",
			api: &KubernetesAPIMock{
				daemonSet: []appsv1.DaemonSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: now.Add(-10 * time.Hour),
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							CreationTimestamp: metav1.Time{
								Time: now.Add(-5 * time.Hour),
							},
						},
					},
				},
			},
			expectedAge: ResourceAge(18000),
			found:       true,
			wantErr:     false,
		},
		{
			name: "empty daemonset list",
			api: &KubernetesAPIMock{
				daemonSet: []appsv1.DaemonSet{},
			},
			expectedAge: 0,
			found:       false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			cancel()

			got, found, err := YoungestDaemonsetAge(ctx, tt.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoungestDaemonsetAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedAge {
				t.Errorf("YoungestDaemonsetAge() = %v, want %v", got, tt.expectedAge)
			}
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
		})
	}
}
