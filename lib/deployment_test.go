package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestYoungestDeploymentAge(t *testing.T) {
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
				err: errors.New("pseudo random k8s appsv1 deployments list error"),
			},
			expectedAge: 0,
			found:       false,
			wantErr:     true,
		},
		{
			name: "get correct deployment age 10h",
			api: &KubernetesAPIMock{
				deployments: []appsv1.Deployment{
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
			name: "get correct deployment age 5h",
			api: &KubernetesAPIMock{
				deployments: []appsv1.Deployment{
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
			name: "empty deployment list",
			api: &KubernetesAPIMock{
				deployments: []appsv1.Deployment{},
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

			got, found, err := YoungestDeploymentAge(ctx, tt.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoungestDeploymentAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedAge {
				t.Errorf("YoungestDeploymentAge() = %v, want %v", got, tt.expectedAge)
			}
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
		})
	}
}
