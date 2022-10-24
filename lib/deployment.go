package gc

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func YoungestDeploymentAge(ctx context.Context, api KubernetesAPI) (ResourceAge, bool, error) {
	deployments, err := api.Deployments(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("unable to list deployments (k8s appsv1 deployment client): %v", err)
	}

	creationTimestampGetter := func(item appsv1.Deployment) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(deployments, creationTimestampGetter)
}
