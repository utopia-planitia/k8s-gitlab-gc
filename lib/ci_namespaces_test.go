package gc

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

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
			s, err := isHashbased(tt.in)
			if err != nil {
				t.Errorf("isHashbased (%s) returned an error", tt.in)
			}
			if s != tt.out {
				t.Errorf("isHashbased (%s) => %t want %t", tt.in, s, tt.out)
			}
		})
	}
}

type TypedCoreV1Client_mock struct { //mock CoreV1Interface
	namespaces typedcorev1.NamespaceInterface //add setter func?
	pods       typedcorev1.PodInterface       //add setter func?
}

func (c *TypedCoreV1Client_mock) RESTClient() rest.Interface {
	panic("mocked RESTClient not implemented")
}
func (c *TypedCoreV1Client_mock) ComponentStatuses() typedcorev1.ComponentStatusInterface {
	panic("mocked ComponentStatuses not implemented")
}
func (c *TypedCoreV1Client_mock) ConfigMaps(namespace string) typedcorev1.ConfigMapInterface {
	panic("mocked ConfigMaps not implemented")
}
func (c *TypedCoreV1Client_mock) Endpoints(namespace string) typedcorev1.EndpointsInterface {
	panic("mocked Endpoints not implemented")
}
func (c *TypedCoreV1Client_mock) Events(namespace string) typedcorev1.EventInterface {
	panic("mocked Events not implemented")
}
func (c *TypedCoreV1Client_mock) LimitRanges(namespace string) typedcorev1.LimitRangeInterface {
	panic("mocked LimitRanges not implemented")
}
func (c *TypedCoreV1Client_mock) Namespaces() typedcorev1.NamespaceInterface {
	return c.namespaces
}
func (c *TypedCoreV1Client_mock) Nodes() typedcorev1.NodeInterface {
	panic("mocked Nodes not implemented")
}
func (c *TypedCoreV1Client_mock) PersistentVolumes() typedcorev1.PersistentVolumeInterface {
	panic("mocked PersistentVolumes not implemented")
}
func (c *TypedCoreV1Client_mock) PersistentVolumeClaims(namespace string) typedcorev1.PersistentVolumeClaimInterface {
	panic("mocked PersistentVolumeClaims not implemented")
}
func (c *TypedCoreV1Client_mock) Pods(namespace string) typedcorev1.PodInterface {
	return c.pods
}
func (c *TypedCoreV1Client_mock) PodTemplates(namespace string) typedcorev1.PodTemplateInterface {
	panic("mocked PodTemplates not implemented")
}
func (c *TypedCoreV1Client_mock) ReplicationControllers(namespace string) typedcorev1.ReplicationControllerInterface {
	panic("mocked ReplicationControllers not implemented")
}
func (c *TypedCoreV1Client_mock) ResourceQuotas(namespace string) typedcorev1.ResourceQuotaInterface {
	panic("mocked ResourceQuotas not implemented")
}
func (c *TypedCoreV1Client_mock) Secrets(namespace string) typedcorev1.SecretInterface {
	panic("mocked Secrets not implemented")
}
func (c *TypedCoreV1Client_mock) Services(namespace string) typedcorev1.ServiceInterface {
	panic("mocked Services not implemented")
}
func (c *TypedCoreV1Client_mock) ServiceAccounts(namespace string) typedcorev1.ServiceAccountInterface {
	panic("mocked ServiceAccounts not implemented")
}

type namespaces_mock struct {
	list      *v1.NamespaceList
	deletions int
}

func (c *namespaces_mock) Create(ctx context.Context, namespace *v1.Namespace, opts metav1.CreateOptions) (*v1.Namespace, error) {
	panic("mocked Create not implemented")
}
func (c *namespaces_mock) Update(ctx context.Context, namespace *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error) {
	panic("mocked Update not implemented")
}
func (c *namespaces_mock) UpdateStatus(ctx context.Context, namespace *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error) {
	panic("mocked UpdateStatus not implemented")
}
func (c *namespaces_mock) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	c.deletions++
	return nil
}
func (c *namespaces_mock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Namespace, error) {
	panic("mocked Get not implemented")
}
func (c *namespaces_mock) List(ctx context.Context, opts metav1.ListOptions) (*v1.NamespaceList, error) {
	return c.list, nil
}
func (c *namespaces_mock) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	panic("mocked Watch not implemented")
}
func (c *namespaces_mock) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Namespace, err error) {
	panic("mocked Patch not implemented")
}
func (c *namespaces_mock) Apply(ctx context.Context, namespace *corev1.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Namespace, err error) {
	panic("mocked Apply not implemented")
}
func (c *namespaces_mock) ApplyStatus(ctx context.Context, namespace *corev1.NamespaceApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Namespace, err error) {
	panic("mocked ApplyStatus not implemented")
}
func (c *namespaces_mock) Finalize(ctx context.Context, item *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error) {
	panic("mocked Finalize not implemented")
}

type pods_mock struct {
	list            *v1.PodList
	returnListError error
}

func (c *pods_mock) Create(ctx context.Context, pod *v1.Pod, opts metav1.CreateOptions) (*v1.Pod, error) {
	panic("mocked Create not implemented")
}
func (c *pods_mock) Update(ctx context.Context, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error) {
	panic("mocked Update not implemented")
}
func (c *pods_mock) UpdateStatus(ctx context.Context, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error) {
	panic("mocked UpdateStatus not implemented")
}
func (c *pods_mock) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	panic("mocked Delete not implemented")
}
func (c *pods_mock) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	panic("mocked DeleteCollection not implemented")
}
func (c *pods_mock) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	panic("mocked Get not implemented")
}
func (c *pods_mock) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	return c.list, c.returnListError
}
func (c *pods_mock) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	panic("mocked Watch not implemented")
}
func (c *pods_mock) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Pod, err error) {
	panic("mocked Patch not implemented")
}
func (c *pods_mock) Apply(ctx context.Context, pod *corev1.PodApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Pod, err error) {
	panic("mocked Apply not implemented")
}
func (c *pods_mock) ApplyStatus(ctx context.Context, pod *corev1.PodApplyConfiguration, opts metav1.ApplyOptions) (result *v1.Pod, err error) {
	panic("mocked ApplyStatus not implemented")
}
func (c *pods_mock) UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error) {
	panic("mocked UpdateEphemeralContainers not implemented")
}

//	PodExpansion
func (c *pods_mock) Bind(ctx context.Context, binding *v1.Binding, opts metav1.CreateOptions) error {
	panic("mocked Bind not implemented")
}
func (c *pods_mock) Evict(ctx context.Context, eviction *policyv1beta1.Eviction) error {
	panic("mocked Evict not implemented")
}
func (c *pods_mock) EvictV1(ctx context.Context, eviction *policyv1.Eviction) error {
	panic("mocked EvictV1 not implemented")
}
func (c *pods_mock) EvictV1beta1(ctx context.Context, eviction *policyv1beta1.Eviction) error {
	panic("mocked EvictV1beta1 not implemented")
}
func (c *pods_mock) GetLogs(name string, opts *v1.PodLogOptions) *rest.Request {
	panic("mocked GetLogs not implemented")
}
func (c *pods_mock) ProxyGet(scheme, name, port, path string, params map[string]string) rest.ResponseWrapper {
	panic("mocked ProxyGet not implemented")
}

func Test_ContinuousIntegrationNamespaces(t *testing.T) {
	type args struct {
		k8sClients KubernetesClients
		// k8sCoreClient     *TypedCoreV1Client_mock
		ageFuncs          []YoungestResourceAgeFunc
		expectedDeletes   int
		protectedBranches []string
		optOutAnnotations []string
		maxTestingAge     int64
		maxReviewAge      int64
		dryRun            bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "keep namespace",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{ObjectMeta: metav1.ObjectMeta{
									Name: "testing",
									CreationTimestamp: metav1.NewTime(
										time.Now(),
									),
								}},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now(),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
		{
			name: "keep ci namespace",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{ObjectMeta: metav1.ObjectMeta{
									Name: "testing-ci",
									CreationTimestamp: metav1.Time{
										Time: time.Now().Add(-1 * time.Hour),
									},
								}},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now().Add(-1 * time.Hour),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
		{
			name: "delete ci namespace",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{ObjectMeta: metav1.ObjectMeta{
									Name: "ci-testing-d41d8cd98f00b204e9800998ecf8427e",
									CreationTimestamp: metav1.Time{
										Time: time.Now().Add(-10 * time.Hour),
									},
								}},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now().Add(-10 * time.Hour),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   1,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
		{
			name: "skip terminating namespace",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{
									ObjectMeta: metav1.ObjectMeta{
										Name: "ci-terminating-d41d8cd98f00b204e9800998ecf8427e",
										CreationTimestamp: metav1.Time{
											Time: time.Now().Add(-10 * time.Hour),
										},
									},
									Status: v1.NamespaceStatus{
										Phase: v1.NamespaceTerminating,
									},
								},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now().Add(-10 * time.Hour),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
		{
			name: "it should delete ns - when ns age implies deletion + pod age is (almost) same as ns age",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{ObjectMeta: metav1.ObjectMeta{
									Name: "ci-testing-d41d8cd98f00b204e9800998ecf8427e",
									CreationTimestamp: metav1.Time{
										Time: time.Now().Add(-10 * time.Hour),
									},
								}},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now().Add(-10 * time.Hour),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   1,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
		//it should not delete ns - when ns age implies deletion but pod age is to young
		{
			name: "it should not delete ns - when ns age implies deletion but pod age is to young",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{ObjectMeta: metav1.ObjectMeta{
									Name: "ci-testing-d41d8cd98f00b204e9800998ecf8427e",
									CreationTimestamp: metav1.Time{
										Time: time.Now().Add(-10 * time.Hour),
									},
								}},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now().Add(-1 * time.Hour),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
		//it should delete ns - when ns implies deletion & pod age implies deletion
		{
			name: "it should delete ns - when ns implies deletion & pod age implies deletion",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						namespaces: &namespaces_mock{
							list: &v1.NamespaceList{Items: []v1.Namespace{
								{ObjectMeta: metav1.ObjectMeta{
									Name: "ci-testing-d41d8cd98f00b204e9800998ecf8427e",
									CreationTimestamp: metav1.Time{
										Time: time.Now().Add(-15 * time.Hour),
									},
								}},
							}},
						},
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											CreationTimestamp: metav1.NewTime(
												time.Now().Add(-10 * time.Hour),
											),
										},
									},
								},
							},
						},
					},
				},
				ageFuncs: []YoungestResourceAgeFunc{
					NamespaceAge,
					YoungestPodAge,
				},
				expectedDeletes:   1,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
				dryRun:            false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := ContinuousIntegrationNamespaces(ctx, tt.args.k8sClients, tt.args.ageFuncs, tt.args.protectedBranches, tt.args.optOutAnnotations, tt.args.maxTestingAge, tt.args.maxReviewAge, tt.args.dryRun); (err != nil) != tt.wantErr {
				t.Errorf("ContinuousIntegrationNamespaces() error = %v, wantErr %v", err, tt.wantErr)
			}
			//deletions := tt.args.k8sCoreClient.namespaces.deletions
			deletions := tt.args.k8sClients.CoreV1.(*TypedCoreV1Client_mock).namespaces.(*namespaces_mock).deletions
			if tt.args.expectedDeletes != deletions {
				t.Errorf("deletions = %v, want %v", deletions, tt.args.expectedDeletes)
			}
		})
	}
}

func Test_ageFns_youngestAge(t *testing.T) {
	type args struct {
		k8sClients KubernetesClients
		namespace  v1.Namespace
	}
	tests := []struct {
		name              string
		ageFuncs          []YoungestResourceAgeFunc
		args              args
		want              ResourceAge
		wantErr           bool
		expectedErrorType error
	}{
		{
			name:              "empty ageFns list",
			ageFuncs:          []YoungestResourceAgeFunc{},
			args:              args{k8sClients: KubernetesClients{}, namespace: v1.Namespace{}},
			want:              ResourceAge(0),
			wantErr:           true,
			expectedErrorType: ErrEmptyFnList,
		},
		{
			name: "fn returning NO_AGES_ERROR (e.g. like pod only - empty list) ",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesClients, n v1.Namespace) (ResourceAge, error) {
					return ResourceAge(0), ErrEmptyK8sResourceList
				},
			},
			args: args{
				k8sClients: KubernetesClients{},
				namespace:  v1.Namespace{},
			},
			want:              ResourceAge(0),
			wantErr:           true,
			expectedErrorType: ErrNoAges,
		},
		{
			name: "single function returning age (like ns only, no pods OR only one pod, no ns)",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesClients, n v1.Namespace) (ResourceAge, error) {
					return ResourceAge(54000), nil
				},
			},
			args: args{
				k8sClients: KubernetesClients{},
				namespace:  v1.Namespace{},
			},
			want:              ResourceAge(54000),
			wantErr:           false,
			expectedErrorType: nil,
		},
		{
			name: "two fns, first returns younger age",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesClients, n v1.Namespace) (ResourceAge, error) {
					return ResourceAge(1), nil
				},
				func(c context.Context, k KubernetesClients, n v1.Namespace) (ResourceAge, error) {
					return ResourceAge(2), nil
				},
			},
			args: args{
				k8sClients: KubernetesClients{},
				namespace:  v1.Namespace{},
			},
			want:              ResourceAge(1),
			wantErr:           false,
			expectedErrorType: nil,
		},
		{
			name: "two fns, second returns younger age",
			ageFuncs: []YoungestResourceAgeFunc{
				func(c context.Context, k KubernetesClients, n v1.Namespace) (ResourceAge, error) {
					return ResourceAge(2), nil
				},
				func(c context.Context, k KubernetesClients, n v1.Namespace) (ResourceAge, error) {
					return ResourceAge(1), nil
				},
			},
			args: args{
				k8sClients: KubernetesClients{},
				namespace:  v1.Namespace{},
			},
			want:              ResourceAge(1),
			wantErr:           false,
			expectedErrorType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, err := youngestAge(ctx, tt.ageFuncs, tt.args.k8sClients, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ageFns.youngestAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ageFns.youngestAge() = %v, want %v", got, tt.want)
			}
			if tt.wantErr {
				if !errors.Is(err, tt.expectedErrorType) {
					t.Errorf(
						"got error type = %v with value \"%s\", expected error type = %v",
						reflect.TypeOf(err), reflect.ValueOf(err), reflect.TypeOf(tt.expectedErrorType),
					)
					return
				}
			}
		})
	}
}

func Test_namespaceAge(t *testing.T) {
	now := time.Now()
	type args struct {
		k8sClients  KubernetesClients
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
				k8sClients: KubernetesClients{},
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: "testing",
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
				k8sClients: KubernetesClients{},
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: "testing",
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

			namespace_age, err := NamespaceAge(ctx, tt.args.k8sClients, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("namespaceAge() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.args.expectedAge != namespace_age {
				t.Errorf("Namespace age = %v, want %v", namespace_age, tt.args.expectedAge)
			}
		})
	}
}

func Test_youngestPodAge(t *testing.T) {
	now := time.Now()
	type args struct {
		k8sClients KubernetesClients
		namespace  v1.Namespace
	}
	tests := []struct {
		name              string
		args              args
		expectedAge       ResourceAge
		wantErr           bool
		checkErrorType    bool
		expectedErrorType error
	}{
		{
			name: "get correct pod age 10h",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											Name: "testingPod",
											CreationTimestamp: metav1.Time{
												Time: now.Add(-10 * time.Hour),
											},
										},
									},
								},
							},
						},
					},
				},
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: "testing",
				}},
			},
			expectedAge:       ResourceAge(36000),
			wantErr:           false,
			checkErrorType:    false,
			expectedErrorType: nil,
		},
		{
			name: "get correct pod age 5h",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{
									{
										ObjectMeta: metav1.ObjectMeta{
											Name: "testingPod",
											CreationTimestamp: metav1.Time{
												Time: now.Add(-5 * time.Hour),
											},
										},
									},
								},
							},
						},
					},
				},
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: "testing",
				}},
			},
			expectedAge:       ResourceAge(18000),
			wantErr:           false,
			checkErrorType:    false,
			expectedErrorType: nil,
		},
		{
			name: "empty pod list - expect error",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{},
							},
						},
					},
				},
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: "testing",
				}},
			},
			expectedAge:       ResourceAge(-1),
			wantErr:           true,
			checkErrorType:    true,
			expectedErrorType: ErrEmptyK8sResourceList,
		},
		{
			name: "expect list error (from k8s client side)",
			args: args{
				k8sClients: KubernetesClients{
					CoreV1: &TypedCoreV1Client_mock{
						pods: &pods_mock{
							list: &v1.PodList{
								Items: []v1.Pod{},
							},
							returnListError: errors.New("pseudo random k8s appsv1 pods list error"),
						},
					},
				},
				namespace: v1.Namespace{ObjectMeta: metav1.ObjectMeta{
					Name: "testing",
				}},
			},
			expectedAge:       ResourceAge(-1),
			wantErr:           true,
			checkErrorType:    false,
			expectedErrorType: nil,
		},

		//todo: test podList with multiple pods
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			got, err := YoungestPodAge(ctx, tt.args.k8sClients, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("youngestPodAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.checkErrorType {
				if !errors.Is(err, tt.expectedErrorType) {
					t.Errorf(
						"got error type = %v, expected error type = %v, got err msg=\"%s\", expected err msg=\"%s\"",
						reflect.TypeOf(err), reflect.TypeOf(tt.expectedErrorType), err, tt.expectedErrorType,
					)
					return
				}
			}

			if got != tt.expectedAge {
				t.Errorf("youngestPodAge() = %v, want %v", got, tt.expectedAge)
			}
		})
	}
}
