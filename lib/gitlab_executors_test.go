package gc

import "testing"

func Test_isGitlabJobPod(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{
			name:   "empty map",
			labels: map[string]string{},
			want:   false,
		},
		{
			name: "is other pod",
			labels: map[string]string{
				"a": "b",
			},
			want: false,
		},
		{
			name: "is runner pod",
			labels: map[string]string{
				"app": "gitlab-ci-job",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGitlabJobPod(tt.labels); got != tt.want {
				t.Errorf("isGitlabJobPod() = %v, want %v", got, tt.want)
			}
		})
	}
}
