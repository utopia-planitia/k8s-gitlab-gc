package gc

import (
	"fmt"
	"time"

	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const twoHours = 60 * 60 * 2

// GitlabExecutors removes gitlab execution pods with an age above 2 hours
func GitlabExecutors(client corev1.PodInterface) error {

	pods, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return nil
	}

	for _, pod := range pods.Items {
		if isOld(pod) && belongsToProject(pod) {
			fmt.Printf("deleting Pod %s\n", pod.ObjectMeta.Name)
			client.Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{})
		}
	}

	return nil
}

func isOld(pod v1.Pod) bool {
	age := time.Now().Unix() - pod.Status.StartTime.Unix()
	return age > twoHours
}

func belongsToProject(pod v1.Pod) bool {
	return strings.Contains(pod.ObjectMeta.Name, "-project-")
}
