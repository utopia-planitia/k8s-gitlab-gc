package gc

import (
	"context"
	"math"
)

func YoungestPodAge(ctx context.Context, api KubernetesAPI) (ResourceAge, bool, error) {
	pods, err := api.Pods(ctx)
	if err != nil {
		return 0, false, err
	}

	if len(pods) == 0 {
		return 0, false, nil
	}

	youngestResourceAge := int64(math.MaxInt64)
	for _, pod := range pods {
		age := age(pod.ObjectMeta.CreationTimestamp)

		if age < int64(youngestResourceAge) {
			youngestResourceAge = age
		}
	}

	return ResourceAge(youngestResourceAge), true, nil
}
