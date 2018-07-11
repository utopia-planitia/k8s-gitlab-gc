package gc

import (
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// GitlabExecutors removes gitlab execution pods with an age above 2 hours
func GitlabExecutors(client corev1.PodInterface, maxAge int64) error {

	pods, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return nil
	}

	for _, pod := range pods.Items {

		if !isTaggedBy(pod.ObjectMeta.Name, "project") {
			continue
		}

		if age(pod.ObjectMeta.CreationTimestamp) < maxAge {
			continue
		}

		fmt.Printf("deleting pod %s\n", pod.ObjectMeta.Name)
		client.Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{})
	}

	return nil
}

func isTaggedBy(s, t string) bool {
	return strings.HasPrefix(s, t+"-") || strings.Contains(s, "-"+t+"-") || strings.HasSuffix(s, "-"+t)
}

func age(t metav1.Time) int64 {
	return time.Now().Unix() - t.Unix()
}
