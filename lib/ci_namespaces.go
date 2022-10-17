package gc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type KubernetesClients struct {
	CoreV1  corev1.CoreV1Interface
	AppsV1  typedappsv1.AppsV1Interface
	BatchV1 typedbatchv1.BatchV1Interface
}

type ResourceAge int64
type YoungestResourceAgeFunc func(ctx context.Context, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error)

var ErrEmptyK8sResourceList error = errors.New("emptyK8sResourceList")
var ErrEmptyFnList error = errors.New("no ageFn functions in list provided")
var ErrNoAges error = errors.New("couldn't get a single resource age from ageFns")

func youngestAge(ctx context.Context, ageFuncs []YoungestResourceAgeFunc, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	if len(ageFuncs) == 0 {
		return ResourceAge(0), ErrEmptyFnList
	}

	ages := []ResourceAge{}
	for _, ageFn := range ageFuncs {
		age, err := ageFn(ctx, k8sClients, namespace)
		if err != nil {
			// when we got a namespace with no pods
			if errors.Is(err, ErrEmptyK8sResourceList) {
				continue
			}
			return -1, fmt.Errorf("%s", err)
		}

		ages = append(ages, age)
	}

	if len(ages) == 0 {
		return ResourceAge(0), ErrNoAges
	}

	youngestResourceAge := ages[0]
	for _, age := range ages {
		if age < youngestResourceAge {
			youngestResourceAge = age
		}
	}

	return ResourceAge(youngestResourceAge), nil
}

// ContinuousIntegrationNamespaces removes no longer used namespaces
func ContinuousIntegrationNamespaces(ctx context.Context, k8sClients KubernetesClients, ageFuncs []YoungestResourceAgeFunc, protectedBranches, optOutAnnotations []string, maxTestingAge, maxReviewAge int64, dryRun bool) error {
	corev1Client := k8sClients.CoreV1
	namespaces := corev1Client.Namespaces()
	nss, err := namespaces.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(nss.Items) == 0 {
		return nil
	}

	for _, ns := range nss.Items {
		if isTerminating(ns) {
			continue
		}

		name := ns.ObjectMeta.Name

		if isProtected(name, protectedBranches) {
			continue
		}

		if !isCI(name) {
			continue
		}

		if hasOptedOut(ns.ObjectMeta.Annotations, optOutAnnotations) {
			continue
		}

		isHashbased, err := isHashbased(name)
		if err != nil {
			return fmt.Errorf("failed to check for hash based namespace '%s': %s", name, err)
		}

		maxAge := maxReviewAge
		if isHashbased {
			maxAge = maxTestingAge
		}

		age, err := youngestAge(ctx, ageFuncs, k8sClients, ns)
		if err != nil {
			return err
		}

		if int64(age) < maxAge {
			continue
		}

		fmt.Printf("deleting namespace: %s, age: %d, maxAge: %d, ageInHours: %d, ageInDays: %d\n", name, age, maxAge, age/60/60, age/60/60/24)

		if dryRun {
			continue
		}

		err = namespaces.Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func NamespaceAge(_ context.Context, _ KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	return ResourceAge(age(namespace.ObjectMeta.CreationTimestamp)), nil
}

func YoungestPodAge(ctx context.Context, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	podsClient := k8sClients.CoreV1.Pods(namespace.ObjectMeta.Name)
	pods, err := podsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return -1, err
	}

	if len(pods.Items) == 0 {
		return -1, ErrEmptyK8sResourceList
	}

	youngestResourceAge := int64(math.MaxInt64)
	for _, pod := range pods.Items {
		age := age(pod.ObjectMeta.CreationTimestamp)

		if age < int64(youngestResourceAge) {
			youngestResourceAge = age
		}
	}

	return ResourceAge(youngestResourceAge), nil
}

func YoungestDeploymentAge(ctx context.Context, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	deploymentClient := k8sClients.AppsV1.Deployments(namespace.ObjectMeta.Name)
	deployments, err := deploymentClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return -1, fmt.Errorf("unable to list deployments (k8s appsv1 deployment client): %s", err)
	}

	creationTimestampGetter := func(item appsv1.Deployment) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(deployments.Items, creationTimestampGetter)
}

func YoungestStatefulsetAge(ctx context.Context, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	statefulsetClient := k8sClients.AppsV1.StatefulSets(namespace.ObjectMeta.Name)
	statefulsets, err := statefulsetClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return -1, fmt.Errorf("unable to list statefulsets (k8s appsv1 statefulset client): %s", err)
	}

	creationTimestampGetter := func(item appsv1.StatefulSet) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(statefulsets.Items, creationTimestampGetter)
}

func YoungestDaemonsetAge(ctx context.Context, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	daemonsetsClient := k8sClients.AppsV1.DaemonSets(namespace.ObjectMeta.Name)
	daemonsets, err := daemonsetsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return -1, fmt.Errorf("unable to list daemonsets (k8s appsv1 daemonset client): %s", err)
	}

	creationTimestampGetter := func(item appsv1.DaemonSet) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(daemonsets.Items, creationTimestampGetter)
}

func YoungestCronjobAge(ctx context.Context, k8sClients KubernetesClients, namespace v1.Namespace) (ResourceAge, error) {
	cronjobsClient := k8sClients.BatchV1.CronJobs(namespace.ObjectMeta.Name)
	cronjobs, err := cronjobsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return -1, fmt.Errorf("unable to list cronjobs (k8s appsv1 cronjob client): %s", err)
	}

	creationTimestampGetter := func(item batchv1.CronJob) metav1.Time {
		return item.ObjectMeta.CreationTimestamp
	}

	return getYoungestItemsResourceAge(cronjobs.Items, creationTimestampGetter)
}

func getYoungestItemsResourceAge[item any](items []item, creationTimestampGetter func(item) metav1.Time) (ResourceAge, error) {
	if len(items) == 0 {
		return -1, ErrEmptyK8sResourceList
	}

	youngestResourceAge := age(creationTimestampGetter(items[0]))
	for _, item := range items {
		age := age(creationTimestampGetter(item))

		if age < int64(youngestResourceAge) {
			youngestResourceAge = age
		}
	}

	return ResourceAge(youngestResourceAge), nil
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
