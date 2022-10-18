package gc

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getYoungestItemsResourceAge(t *testing.T) {
	now := time.Now()

	type testType struct {
		ts metav1.Time
	}

	type args struct {
		Items                   []testType
		creationTimestampGetter func(testType) metav1.Time
	}

	tests := []struct {
		name        string
		args        args
		expectedAge ResourceAge
		found       bool
		wantErr     bool
	}{
		{
			name: "get correct age 5h",
			args: args{
				Items: []testType{
					{
						ts: metav1.NewTime(now.Add(-5 * time.Hour)),
					},
					{
						ts: metav1.NewTime(now.Add(-10 * time.Hour)),
					},
				},
				creationTimestampGetter: func(item testType) metav1.Time {
					return item.ts
				},
			},
			expectedAge: ResourceAge(18000),
			found:       true,
			wantErr:     false,
		},
		{
			name: "empty list",
			args: args{
				Items: []testType{},
				creationTimestampGetter: func(item testType) metav1.Time {
					return item.ts
				},
			},
			expectedAge: 0,
			found:       false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, found, err := getYoungestItemsResourceAge(tt.args.Items, tt.args.creationTimestampGetter)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoungestItemsResourceAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expectedAge {
				t.Errorf("YoungestItemsResourceAge() = %v, want %v", got, tt.expectedAge)
			}
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
		})
	}
}

func Test_youngestAge(t *testing.T) {
	tests := []struct {
		name     string
		ageFuncs []YoungestResourceAgeFunc
		want     ResourceAge
		found    bool
		wantErr  bool
	}{
		{
			name:     "empty ageFns list",
			ageFuncs: []YoungestResourceAgeFunc{},
			want:     ResourceAge(0),
			found:    false,
			wantErr:  false,
		},
		{
			name: "fn returning NO_AGES_ERROR (e.g. like pod only - empty list) ",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesAPI) (ResourceAge, bool, error) {
					return ResourceAge(0), false, nil
				},
			},
			want:    ResourceAge(0),
			found:   false,
			wantErr: false,
		},
		{
			name: "single function returning age (like ns only, no pods OR only one pod, no ns)",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesAPI) (ResourceAge, bool, error) {
					return ResourceAge(54000), true, nil
				},
			},
			want:    ResourceAge(54000),
			found:   true,
			wantErr: false,
		},
		{
			name: "two fns, first returns younger age",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesAPI) (ResourceAge, bool, error) {
					return ResourceAge(1), true, nil
				},
				func(c context.Context, k KubernetesAPI) (ResourceAge, bool, error) {
					return ResourceAge(2), true, nil
				},
			},
			want:    ResourceAge(1),
			found:   true,
			wantErr: false,
		},
		{
			name: "two fns, second returns younger age",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesAPI) (ResourceAge, bool, error) {
					return ResourceAge(2), true, nil
				},
				func(c context.Context, k KubernetesAPI) (ResourceAge, bool, error) {
					return ResourceAge(1), true, nil
				},
			},
			want:    ResourceAge(1),
			found:   true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, found, err := youngestAge(ctx, tt.ageFuncs, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("ageFns.youngestAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ageFns.youngestAge() = %v, want %v", got, tt.want)
			}
			if found != tt.found {
				t.Errorf("found = %v, want %v", found, tt.found)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("ageFns.youngestAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
