package gc

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// GitlabExecutors removes gitlab execution pods with an age above 2 hours
func GitlabExecutors(ctx context.Context, client corev1.PodInterface, maxAge int64) error {

	pods, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return nil
	}

	for _, pod := range pods.Items {

		name := pod.ObjectMeta.Name

		if !isTaggedBy(name, "project") {
			continue
		}

		age := age(pod.ObjectMeta.CreationTimestamp)
		if age < maxAge {
			continue
		}

		fmt.Printf("deleting pod: %s, age: %d, maxAge: %d, ageInHours: %d\n", name, age, maxAge, age/60/60)
		err = client.Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func isTaggedBy(s, t string) bool {
	return strings.HasPrefix(s, t+"-") || strings.Contains(s, "-"+t+"-") || strings.HasSuffix(s, "-"+t)
}

func age(t metav1.Time) int64 {
	return time.Now().Unix() - t.Unix()
}
