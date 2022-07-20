package gc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type resourceAge int64
type youngestResourceAgeFunc func(ctx context.Context, k8sCoreClient corev1.CoreV1Interface, namespace v1.Namespace) (resourceAge, error)

var ErrEmptyK8sResourceList error = errors.New("emptyK8sResourceList")
var ErrEmptyFnList error = errors.New("no ageFn functions in list provided")
var ErrNoAges error = errors.New("couldn't get a single resource age from ageFns")

func getDefaultResourceAgeFuncs() []youngestResourceAgeFunc {
	return []youngestResourceAgeFunc{
		namespaceAge,
		youngestPodAge,
	}
}

func youngestAge(ctx context.Context, ageFuncs []youngestResourceAgeFunc, k8sCoreClient corev1.CoreV1Interface, namespace v1.Namespace) (resourceAge, error) {
	if len(ageFuncs) == 0 {
		return resourceAge(0), ErrEmptyFnList
	}

	ages := []resourceAge{}
	for _, ageFn := range ageFuncs {
		age, err := ageFn(ctx, k8sCoreClient, namespace)
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
		return resourceAge(0), ErrNoAges
	}

	youngestResourceAge := ages[0]
	for _, age := range ages {
		if age < youngestResourceAge {
			youngestResourceAge = age
		}
	}

	return resourceAge(youngestResourceAge), nil
}

// ContinuousIntegrationNamespaces removes no longer used namespaces
func ContinuousIntegrationNamespaces(ctx context.Context, k8sCoreClient corev1.CoreV1Interface, protectedBranches, optOutAnnotations []string, maxTestingAge, maxReviewAge int64) error {
	namespaces := k8sCoreClient.Namespaces()
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

		age, err := youngestAge(ctx, getDefaultResourceAgeFuncs(), k8sCoreClient, ns)
		if err != nil {
			return err
		}

		if int64(age) < maxAge {
			continue
		}

		fmt.Printf("deleting namespace: %s, age: %d, maxAge: %d, ageInHours: %d, ageInDays: %d\n", name, age, maxAge, age/60/60, age/60/60/24)
		err = namespaces.Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func namespaceAge(ctx context.Context, k8sCoreClient corev1.CoreV1Interface, namespace v1.Namespace) (resourceAge, error) {
	return resourceAge(age(namespace.ObjectMeta.CreationTimestamp)), nil
}

func youngestPodAge(ctx context.Context, k8sCoreClient corev1.CoreV1Interface, namespace v1.Namespace) (resourceAge, error) {
	podsClient := k8sCoreClient.Pods(namespace.ObjectMeta.Name)
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

	return resourceAge(youngestResourceAge), nil
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
