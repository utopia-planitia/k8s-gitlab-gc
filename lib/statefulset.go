package gc

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func YoungestStatefulsetAge(ctx context.Context, api KubernetesAPI) (ResourceAge, bool, error) {
	statefulsets, err := api.StatefulSets(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("unable to list statefulsets (k8s appsv1 statefulset client): %v", err)
	}

	creationTimestampGetter := func(item appsv1.StatefulSet) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(statefulsets, creationTimestampGetter)
}
