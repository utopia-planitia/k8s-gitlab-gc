package gc

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func YoungestDaemonsetAge(ctx context.Context, api KubernetesAPI) (ResourceAge, bool, error) {
	daemonsets, err := api.DaemonSets(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("unable to list daemonsets (k8s appsv1 daemonset client): %v", err)
	}

	creationTimestampGetter := func(item appsv1.DaemonSet) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(daemonsets, creationTimestampGetter)
}
