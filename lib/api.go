package gc

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesAPI interface {
	Pods(ctx context.Context) ([]v1.Pod, error)
	Deployments(ctx context.Context) ([]appsv1.Deployment, error)
	StatefulSets(ctx context.Context) ([]appsv1.StatefulSet, error)
	DaemonSets(ctx context.Context) ([]appsv1.DaemonSet, error)
	CronJobs(ctx context.Context) ([]batchv1.CronJob, error)
	Namespace() v1.Namespace
	DeleteCurrentNamespace(ctx context.Context) error
}

type ResourceAge int64
type YoungestResourceAgeFunc func(ctx context.Context, k8sClients KubernetesAPI) (ResourceAge, bool, error)

type KubernetesClient struct {
	clientset *kubernetes.Clientset
	namespace v1.Namespace
}

func (k *KubernetesClient) Pods(ctx context.Context) ([]v1.Pod, error) {
	namespaceName := k.namespace.ObjectMeta.Name
	pods, err := k.clientset.CoreV1().Pods(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (k *KubernetesClient) Deployments(ctx context.Context) ([]appsv1.Deployment, error) {
	namespaceName := k.namespace.ObjectMeta.Name
	deployments, err := k.clientset.AppsV1().Deployments(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return deployments.Items, nil
}

func (k *KubernetesClient) StatefulSets(ctx context.Context) ([]appsv1.StatefulSet, error) {
	namespaceName := k.namespace.ObjectMeta.Name
	statefulSets, err := k.clientset.AppsV1().StatefulSets(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return statefulSets.Items, nil
}

func (k *KubernetesClient) DaemonSets(ctx context.Context) ([]appsv1.DaemonSet, error) {
	namespaceName := k.namespace.ObjectMeta.Name
	daemonSets, err := k.clientset.AppsV1().DaemonSets(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return daemonSets.Items, nil
}

func (k *KubernetesClient) CronJobs(ctx context.Context) ([]batchv1.CronJob, error) {
	namespaceName := k.namespace.ObjectMeta.Name
	cronJobs, err := k.clientset.BatchV1().CronJobs(namespaceName).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return cronJobs.Items, nil
}

func (k *KubernetesClient) Namespace() v1.Namespace {
	return k.namespace
}

func (k *KubernetesClient) DeleteCurrentNamespace(ctx context.Context) error {
	namespaceName := k.namespace.ObjectMeta.Name
	return k.clientset.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
}

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
