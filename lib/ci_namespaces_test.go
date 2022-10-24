package gc

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesAPIMock struct {
	pods             []v1.Pod
	deployments      []appsv1.Deployment
	statefulSet      []appsv1.StatefulSet
	daemonSet        []appsv1.DaemonSet
	cronJobs         []batchv1.CronJob
	namespace        v1.Namespace
	err              error
	namespaceDeleted bool
}

func (k *KubernetesAPIMock) Pods(ctx context.Context) ([]v1.Pod, error) {
	return k.pods, k.err
}

func (k *KubernetesAPIMock) Deployments(ctx context.Context) ([]appsv1.Deployment, error) {
	return k.deployments, k.err
}

func (k *KubernetesAPIMock) StatefulSets(ctx context.Context) ([]appsv1.StatefulSet, error) {
	return k.statefulSet, k.err
}

func (k *KubernetesAPIMock) DaemonSets(ctx context.Context) ([]appsv1.DaemonSet, error) {
	return k.daemonSet, k.err
}

func (k *KubernetesAPIMock) CronJobs(ctx context.Context) ([]batchv1.CronJob, error) {
	return k.cronJobs, k.err
}

func (k *KubernetesAPIMock) Namespace() v1.Namespace {
	return k.namespace
}

func (k *KubernetesAPIMock) DeleteCurrentNamespace(_ context.Context) error {
	k.namespaceDeleted = true
	return nil
}

var isHashbasedTests = []struct {
	in  string
	out bool
}{
	{"project-shop-ci-54823-3a5db1781ab7cde0c53a3b53d995b75ee5873243", true},
	{"project-shop-ci-feature-cloud-upload-service", false},
	{"ci-user-user-feature-upgrade-framework-5-9-21", false},
	{"ci-user-user-feature-tes-562-api-stock-import", false},
	{"ci-username-playground-project-selenium-86421-efaf6b99b3ae85f7a405", true},
	{"ci-username-playground-project-selenium-86421-efaf6b99b3ae85f", true},
	{"ci-username-playground-project-selenium-86421-efaf6b99b3ae85", false},
	{"ci-username-playground-project-selenium-86421-f", false},
}

func TestIsHashbased(t *testing.T) {
	for _, tt := range isHashbasedTests {
		t.Run(tt.in, func(t *testing.T) {
			s := hashRegex.MatchString(tt.in)
			if s != tt.out {
				t.Errorf("isHashbased (%s) => %t want %t", tt.in, s, tt.out)
			}
		})
	}
}

func TestNamespaceAge(t *testing.T) {
	now := time.Now()
	type args struct {
		namespace   v1.Namespace
		expectedAge ResourceAge
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "get correct namespace age 10h",
			args: args{
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(
						now.Add(-10 * time.Hour),
					),
				}},
				expectedAge: ResourceAge(36000),
			},
			wantErr: false,
		},
		{
			name: "get correct namespace age 5h",
			args: args{
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(
						now.Add(-5 * time.Hour),
					),
				}},
				expectedAge: ResourceAge(18000),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			api := &KubernetesAPIMock{
				namespace: tt.args.namespace,
			}

			namespace_age, found, err := NamespaceAge(ctx, api)
			if (err != nil) != tt.wantErr {
				t.Errorf("namespaceAge() error = %v, wantErr %v", err, tt.wantErr)
			}
			if found != true {
				t.Errorf("found = %v, want %v", found, true) // namespaces always exist
			}
			if tt.args.expectedAge != namespace_age {
				t.Errorf("Namespace age = %v, want %v", namespace_age, tt.args.expectedAge)
			}
		})
	}
}

func Test_shouldDeleteNamespace(t *testing.T) {
	type args struct {
		api               KubernetesAPI
		ageFuncs          []YoungestResourceAgeFunc
		protectedBranches []string
		optOutAnnotations []string
		maxTestingAge     int64
		maxReviewAge      int64
	}
	ageFuncs := []YoungestResourceAgeFunc{
		func(_ context.Context, _ KubernetesAPI) (ResourceAge, bool, error) {
			return ResourceAge(15), true, nil
		},
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "keep namespace",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{}},
				},
				ageFuncs:     ageFuncs,
				maxReviewAge: int64(20),
			},
		},
		{
			name: "keep hash based namespace",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
						Name: "project-shop-ci-54823-3a5db1781ab7cde0c53a3b53d995b75ee5873243",
					}},
				},
				ageFuncs:      ageFuncs,
				maxTestingAge: int64(20), // only for hash based
			},
		},
		{
			name: "keep hash based namespace with age between maxTestingAge and maxReviewAge",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
						Name: "project-shop-ci-54823-3a5db1781ab7cde0c53a3b53d995b75ee5873243",
					}},
				},
				ageFuncs:      ageFuncs,
				maxTestingAge: int64(20), // only for hash based
				maxReviewAge:  int64(10),
			},
		},
		{
			name: "delete hash based namespace",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
						Name: "project-shop-ci-54823-3a5db1781ab7cde0c53a3b53d995b75ee5873243",
					}},
				},
				ageFuncs:      ageFuncs,
				maxTestingAge: int64(10), // only for hash based
			},
			want: true,
		},
		{
			name: "keep ci namespace",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
						Name: "project-shop-ci",
					}},
				},
				ageFuncs:     ageFuncs,
				maxReviewAge: int64(20),
			},
		},
		{
			name: "delete ci namespace",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
						Name: "project-shop-ci",
					}},
				},
				ageFuncs:     ageFuncs,
				maxReviewAge: int64(10),
			},
			want: true,
		},
		{
			name: "skip terminating namespace",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-shop-ci",
						},
						Status: v1.NamespaceStatus{
							Phase: v1.NamespaceTerminating,
						},
					},
				},
				ageFuncs:     ageFuncs,
				maxReviewAge: int64(10),
			},
			want: false,
		},
		{
			name: "keep ns - when ns age implies deletion but pod age is to young",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-shop-ci",
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					func(_ context.Context, _ KubernetesAPI) (ResourceAge, bool, error) {
						return ResourceAge(15), true, nil
					},
					func(_ context.Context, _ KubernetesAPI) (ResourceAge, bool, error) {
						return ResourceAge(5), true, nil
					},
				},
				maxReviewAge: int64(10),
			},
			want: false,
		},
		{
			name: "delete ns - when ns implies deletion & pod age implies deletion",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-shop-ci",
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					func(_ context.Context, _ KubernetesAPI) (ResourceAge, bool, error) {
						return ResourceAge(15), true, nil
					},
					func(_ context.Context, _ KubernetesAPI) (ResourceAge, bool, error) {
						return ResourceAge(12), true, nil
					},
				},
				maxReviewAge: int64(10),
			},
			want: true,
		},
		{
			name: "protect branch",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
						Name: "project-bla-shop-ci",
					}},
				},
				ageFuncs:          ageFuncs,
				maxReviewAge:      int64(10),
				protectedBranches: []string{"bla"},
			},
			want: false,
		},
		{
			name: "opt out annotation",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-shop-ci",
							Annotations: map[string]string{
								"bla": "true",
							},
						},
					},
				},
				ageFuncs:     ageFuncs,
				maxReviewAge: int64(10),
				optOutAnnotations: []string{
					"bla",
				},
			},
			want: false,
		},
		{
			name: "opt out annotation false",
			args: args{
				api: &KubernetesAPIMock{
					namespace: v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "project-shop-ci",
							Annotations: map[string]string{
								"bla": "",
							},
						},
					},
				},
				ageFuncs:     ageFuncs,
				maxReviewAge: int64(10),
				optOutAnnotations: []string{
					"bla",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shouldDeleteNamespace(context.TODO(), tt.args.api, tt.args.ageFuncs, tt.args.protectedBranches, tt.args.optOutAnnotations, tt.args.maxTestingAge, tt.args.maxReviewAge)
			if (err != nil) != tt.wantErr {
				t.Errorf("shouldDeleteNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("shouldDeleteNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
