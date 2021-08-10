package gc

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// GitlabExecutors removes gitlab execution pods
func GitlabExecutors(ctx context.Context, client corev1.PodInterface, maxAge int64) error {
	pods, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return nil
	}

	for _, pod := range pods.Items {
		if !isGitlabJobPod(pod.ObjectMeta.Labels) {
			continue
		}

		age := age(pod.ObjectMeta.CreationTimestamp)
		if age < maxAge {
			continue
		}

		fmt.Printf("deleting pod: %s, age: %d, maxAge: %d, ageInHours: %d\n", pod.ObjectMeta.Name, age, maxAge, age/60/60)
		err = client.Delete(ctx, pod.ObjectMeta.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func isGitlabJobPod(labels map[string]string) bool {
	v, ok := labels["app"]
	if !ok {
		return false
	}

	if v != "gitlab-ci-job" {
		return false
	}

	return true
}

func age(t metav1.Time) int64 {
	return time.Now().Unix() - t.Unix()
}
