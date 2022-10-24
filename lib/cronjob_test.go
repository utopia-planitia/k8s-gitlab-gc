package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestYoungestCronjobAge(t *testing.T) {
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
				err: errors.New("pseudo random k8s appsv1 cronjobs list error"),
			},
			expectedAge: 0,
			found:       false,
			wantErr:     true,
		},
		{
			name: "get correct cronjob age 10h",
			api: &KubernetesAPIMock{
				cronJobs: []batchv1.CronJob{
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
			name: "get correct cronjob age 5h",
			api: &KubernetesAPIMock{
				cronJobs: []batchv1.CronJob{
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
			name: "empty cronjob list",
			api: &KubernetesAPIMock{
				cronJobs: []batchv1.CronJob{},
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

			got, found, err := YoungestCronjobAge(ctx, tt.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoungestCronjobAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedAge {
				t.Errorf("YoungestCronjobAge() = %v, want %v", got, tt.expectedAge)
			}
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
		})
	}
}
