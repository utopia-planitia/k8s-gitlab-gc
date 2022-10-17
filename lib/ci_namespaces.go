package gc

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceAge int64
type YoungestResourceAgeFunc func(ctx context.Context, k8sClients KubernetesAPI) (ResourceAge, bool, error)

func youngestAge(ctx context.Context, ageFuncs []YoungestResourceAgeFunc, api KubernetesAPI) (ResourceAge, bool, error) {
	ages := []ResourceAge{}
	for _, ageFn := range ageFuncs {
		age, found, err := ageFn(ctx, api)
		if err != nil {
			return 0, false, err
		}

		if !found {
			continue
		}

		ages = append(ages, age)
	}

	if len(ages) == 0 {
		return 0, false, nil
	}

	youngestResourceAge := ages[0]
	for _, age := range ages {
		if age < youngestResourceAge {
			youngestResourceAge = age
		}
	}

	return ResourceAge(youngestResourceAge), true, nil
}

// ContinuousIntegrationNamespaces removes no longer used namespaces
func ContinuousIntegrationNamespaces(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	ageFuncs []YoungestResourceAgeFunc,
	protectedBranches,
	optOutAnnotations []string,
	maxTestingAge,
	maxReviewAge int64,
	dryRun bool,
) error {
	namespaces := clientset.CoreV1().Namespaces()
	nss, err := namespaces.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, ns := range nss.Items {
		api := &KubernetesClient{
			namespace: ns,
			clientset: clientset,
		}

		delete, err := shouldDeleteNamespace(
			ctx,
			api,
			ageFuncs,
			protectedBranches,
			optOutAnnotations,
			maxTestingAge,
			maxReviewAge,
		)
		if err != nil {
			return err
		}

		if delete {
			name := ns.ObjectMeta.Name

			fmt.Printf("deleting namespace: %s\n", name)

			if dryRun {
				continue
			}

			err := clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func shouldDeleteNamespace(
	ctx context.Context,
	api KubernetesAPI,
	ageFuncs []YoungestResourceAgeFunc,
	protectedBranches,
	optOutAnnotations []string,
	maxTestingAge,
	maxReviewAge int64,
) (bool, error) {
	ns := api.Namespace()

	if isTerminating(ns) {
		return false, nil
	}

	name := ns.ObjectMeta.Name

	if isProtected(name, protectedBranches) {
		return false, nil
	}

	if !isCI(name) {
		return false, nil
	}

	if hasOptedOut(ns.ObjectMeta.Annotations, optOutAnnotations) {
		return false, nil
	}

	isHashbased, err := isHashbased(name)
	if err != nil {
		return false, fmt.Errorf("failed to check for hash based namespace '%s': %v", name, err)
	}

	maxAge := maxReviewAge
	if isHashbased {
		maxAge = maxTestingAge
	}

	age, found, err := youngestAge(ctx, ageFuncs, api)
	if err != nil {
		return false, err
	}

	if !found {
		return false, fmt.Errorf("no item with an age was found - this should not happen")
	}

	if int64(age) < maxAge {
		return false, nil
	}

	return true, nil
}

func NamespaceAge(_ context.Context, api KubernetesAPI) (ResourceAge, bool, error) {
	return ResourceAge(age(api.Namespace().ObjectMeta.CreationTimestamp)), true, nil
}

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

func getYoungestItemsResourceAge[item any](items []item, creationTimestampGetter func(item) metav1.Time) (ResourceAge, bool, error) {
	if len(items) == 0 {
		return 0, false, nil
	}

	youngestResourceAge := age(creationTimestampGetter(items[0]))
	for _, item := range items {
		age := age(creationTimestampGetter(item))

		if age < int64(youngestResourceAge) {
			youngestResourceAge = age
		}
	}

	return ResourceAge(youngestResourceAge), true, nil
}

func hasOptedOut(annotations map[string]string, optOutAnnotations []string) bool {
	for _, optOutAnnotation := range optOutAnnotations {
		optOut, ok := annotations[optOutAnnotation]
		if !ok {
			continue
		}
		return optOut == "true"
	}
	return false
}

func isCI(name string) bool {
	return isTaggedBy(name, "ci")
}

func isHashbased(name string) (bool, error) {
	return regexp.MatchString("[0-9a-fA-F]{15,}$", name)
}

func isProtected(name string, protectedBranches []string) bool {
	for _, branch := range protectedBranches {
		if isTaggedBy(name, branch) {
			return true
		}
	}
	return false
}

func isTaggedBy(s, t string) bool {
	return strings.HasPrefix(s, t+"-") || strings.Contains(s, "-"+t+"-") || strings.HasSuffix(s, "-"+t)
}

func isTerminating(ns v1.Namespace) bool {
	return ns.Status.Phase == v1.NamespaceTerminating
}
