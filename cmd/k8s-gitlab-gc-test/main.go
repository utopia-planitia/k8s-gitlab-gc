package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	log.Printf("kubeconfig: %v\n", *kubeconfig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	k8s, err := provideKubernetesClient(*kubeconfig)
	if err != nil {
		log.Fatalf("failed initilize kubernetes client: %s", err)
	}

	namespaces := k8s.CoreV1().Namespaces()
	nss, err := namespaces.List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if len(nss.Items) == 0 {
		fmt.Println("zero namespaces listed")
		return
	}

	fmt.Println("namespaces:")
	for _, ns := range nss.Items {
		name := ns.ObjectMeta.Name

		fmt.Println(name)
	}

	// err = gc.ContinuousIntegrationNamespaces(ctx, k8s.CoreV1().Namespaces(), strings.Split(*protectedBranches, ","), strings.Split(*optOutAnnotations, ","), *maxBuildNamespaceAge, *maxReviewNamespaceAge)
	// if err != nil {
	// 	log.Fatalf("failed to clean up ci namespaces: %s", err)
	// }
}

func provideKubernetesClient(kubeconfig string) (*kubernetes.Clientset, error) {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubernetes configuration: %s", err)
	}
	return kubernetes.NewForConfig(k8sConfig)
}
