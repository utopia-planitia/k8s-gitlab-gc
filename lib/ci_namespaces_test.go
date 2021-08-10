package gc

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
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

func TestContinuousIntegrationNamespaces(t *testing.T) {
	type args struct {
		namespaces        *namespaces_mock
		expectedDeletes   int
		protectedBranches []string
		optOutAnnotations []string
		maxTestingAge     int64
		maxReviewAge      int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "keep namespace",
			args: args{
				namespaces: &namespaces_mock{
					list: &v1.NamespaceList{Items: []v1.Namespace{
						{ObjectMeta: metav1.ObjectMeta{Name: "testing"}},
					}},
				},
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
			},
			wantErr: false,
		},
		{
			name: "keep ci namespace",
			args: args{
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
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
			},
			wantErr: false,
		},
		{
			name: "delete ci namespace",
			args: args{
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
				expectedDeletes:   1,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
			},
			wantErr: false,
		},
		{
			name: "skip terminating namespace",
			args: args{
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
				expectedDeletes:   0,
				protectedBranches: []string{},
				optOutAnnotations: []string{},
				maxTestingAge:     int64(60 * 60 * 6),
				maxReviewAge:      int64(60 * 60 * 24 * 2),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := ContinuousIntegrationNamespaces(ctx, tt.args.namespaces, tt.args.protectedBranches, tt.args.optOutAnnotations, tt.args.maxTestingAge, tt.args.maxReviewAge); (err != nil) != tt.wantErr {
				t.Errorf("ContinuousIntegrationNamespaces() error = %v, wantErr %v", err, tt.wantErr)
			}
			deletions := tt.args.namespaces.deletions
			if tt.args.expectedDeletes != deletions {
				t.Errorf("deletions = %v, want %v", deletions, tt.args.expectedDeletes)
			}
		})
	}
}
