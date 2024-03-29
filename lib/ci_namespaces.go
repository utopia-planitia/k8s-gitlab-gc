package gc

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var hashRegex = regexp.MustCompile("[0-9a-fA-F]{15,}$")

// ContinuousIntegrationNamespaces removes no longer used namespaces
func ContinuousIntegrationNamespaces(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	ageFuncs []YoungestResourceAgeFunc,
	protectedBranches,
	optOutAnnotations []string,
	ttlAnnotation string,
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
			ttlAnnotation,
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
	ttlAnnotation string,
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

	maxAge, found, err := ttlAnnotationValue(ns.ObjectMeta.Annotations, ttlAnnotation)
	if err != nil {
		return false, err
	}

	if !found {
		isHashbased := hashRegex.MatchString(name)

		maxAge = maxReviewAge
		if isHashbased {
			maxAge = maxTestingAge
		}
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

func ttlAnnotationValue(annotations map[string]string, ttlAnnotation string) (maxAge int64, found bool, err error) {
	var value string
	var duration time.Duration

	value, found = annotations[ttlAnnotation]
	if !found {
		return
	}

	duration, err = time.ParseDuration(value)
	if err != nil {
		return
	}

	maxAge = int64(duration.Seconds())

	return
}

func isCI(name string) bool {
	return isTaggedBy(name, "ci")
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
