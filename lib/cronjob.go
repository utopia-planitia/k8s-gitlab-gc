package gc

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func YoungestCronjobAge(ctx context.Context, api KubernetesAPI) (ResourceAge, bool, error) {
	cronjobs, err := api.CronJobs(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("unable to list cronjobs (k8s appsv1 cronjob client): %v", err)
	}

	creationTimestampGetter := func(item batchv1.CronJob) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(cronjobs, creationTimestampGetter)
}
