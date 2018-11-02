package gc

import (
	"fmt"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// ContinuousIntegrationNamespaces removes no longer used namespaces
func ContinuousIntegrationNamespaces(client corev1.CoreV1Interface, protectedBranches, optOutAnnotations []string, maxTestingAge, maxReviewAge int64) error {

	// TODO: remove ci namespaces if branch is gone
	// TODO: remove ci namespaces if nothing got updated for 2 days (only clean up .*-ci-.* and keep master / stage / develop)

	nss, err := client.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(nss.Items) == 0 {
		return nil
	}

	for _, ns := range nss.Items {
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

		// TODO check for last modified of: secret, configmap, deployment, statefulset, cronjob, service, ingress, pvc
		age := age(ns.ObjectMeta.CreationTimestamp)
		if age < maxAge {
			continue
		}

		fmt.Printf("deleting namespace: %s, age: %d, maxAge: %d, ageInHours: %d, ageInDays: %d\n", name, age, maxAge, age/60/60, age/60/60/24)
		err = client.Namespaces().Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
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
