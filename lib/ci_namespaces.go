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

// ContinuousIntegrationNamespaces removes no longer used namespaces
func ContinuousIntegrationNamespaces(ctx context.Context, namespaces corev1.NamespaceInterface, protectedBranches, optOutAnnotations []string, maxTestingAge, maxReviewAge int64) error {
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

		age := age(ns.ObjectMeta.CreationTimestamp)
		if age < maxAge {
			continue
		}

		fmt.Printf("deleting namespace: %s, age: %d, maxAge: %d, ageInHours: %d, ageInDays: %d\n", name, age, maxAge, age/60/60, age/60/60/24)
		// err = namespaces.Delete(ctx, name, metav1.DeleteOptions{})
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}

// NamespacesFromResourceAge removes no longer used namespaces based on newest resources age
func NamespacesFromResourceAge(ctx context.Context, k8sCoreInterface corev1.CoreV1Interface) error {
	namespaces := k8sCoreInterface.Namespaces()
	nss, err := namespaces.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Amount of namespaces: %d\n", len(nss.Items))
	if len(nss.Items) == 0 {
		return nil
	}

	for _, ns := range nss.Items {
		fmt.Printf("Analyzing namespace: %s\n", ns.ObjectMeta.Name)

		client := k8sCoreInterface.Pods(ns.ObjectMeta.Name)
		pods, err := client.List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		if len(pods.Items) == 0 {
			fmt.Printf("\tNo Resources found.\n")
			continue
		}

		newestResourceAge, err := getNewestResourceAge(ctx, k8sCoreInterface.Pods(ns.ObjectMeta.Name))
		if err != nil {
			return err
		}

		if isTerminating(ns) {
			continue
		}

		for _, pod := range pods.Items {

			age := age(pod.ObjectMeta.CreationTimestamp)
			if age <= newestResourceAge {
				continue
			}

			fmt.Printf("Deleting pod name: %s, age: %d, newestResourceAge: %d, ageInHours: %d\n", pod.ObjectMeta.Name, age, newestResourceAge, age/60/60)
			// err = client.Delete(ctx, pod.ObjectMeta.Name, metav1.DeleteOptions{})
			// if err != nil {
			// 	return err
			// }
		}
	}

	return nil

}

func getNewestResourceAge(ctx context.Context, client corev1.PodInterface) (int64, error) {
	pods, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return -1, (err)
	}

	if len(pods.Items) == 0 {
		return -1, errors.New("No resources found")
	}

	newestResourceAge := int64(math.MaxInt64)
	for _, pod := range pods.Items {

		age := age(pod.ObjectMeta.CreationTimestamp)
		if age > int64(newestResourceAge) {
			continue
		}
		newestResourceAge = age
	}

	fmt.Printf("Newest Age: %d\n", newestResourceAge)

	return newestResourceAge, nil

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
