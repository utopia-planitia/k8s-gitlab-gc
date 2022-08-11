package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	gc "github.com/utopia-planitia/k8s-gitlab-gc/lib"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var dryRun = flag.Bool("dry-run", false, "execute in dry-run mode - no changes will be applied")
	var kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	var gitlabRunnerNamespace = flag.String("gitlabRunnerNamespace", "gitlab-runner", "namespace to remove gitlab executors from")
	var protectedBranches = flag.String("protectedBranches", "develop,master,main,preview,review,stage,staging", "comma separated list of substrings to mark a namespace as protected from deletion")
	var maxGitlabExecutorAge = flag.Int64("maxGitlabExecutorAge", 70*60, "max age for gitlab executor pods in seconds")
	var maxReviewNamespaceAge = flag.Int64("maxReviewNamespaceAge", 60*60*24*2, "max age for review namespaces in seconds")
	var maxBuildNamespaceAge = flag.Int64("maxBuildNamespaceAge", 60*60*2, "max age for e2e testing namespaces in seconds")
	var optOutAnnotations = flag.String("optOutAnnotations", "disable-automatic-garbage-collection", "comma separated list of annotations to protect namespaces from deletion, annotations need to be set to the string 'true'")
	flag.Parse()

	log.Printf("dryRun: %v\n", *dryRun)
	log.Printf("kubeconfig: %v\n", *kubeconfig)
	log.Printf("gitlabRunnerNamespace: %v\n", *gitlabRunnerNamespace)
	log.Printf("protectedBranches: %v\n", *protectedBranches)
	log.Printf("maxGitlabExecutorAge: %v\n", *maxGitlabExecutorAge)
	log.Printf("maxReviewNamespaceAge: %v\n", *maxReviewNamespaceAge)
	log.Printf("maxBuildNamespaceAge: %v\n", *maxBuildNamespaceAge)
	log.Printf("optOutAnnotations: %v\n", *optOutAnnotations)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	k8s, err := provideKubernetesClient(*kubeconfig)
	if err != nil {
		log.Fatalf("failed initilize kubernetes client: %s", err)
	}

	err = gc.GitlabExecutors(ctx, k8s.CoreV1().Pods(*gitlabRunnerNamespace), *maxGitlabExecutorAge, *dryRun)
	if err != nil {
		log.Fatalf("failed to clean up gitlab executors: %s", err)
	}

	k8sClients := gc.KubernetesClients{
		CoreV1:  k8s.CoreV1(),
		AppsV1:  k8s.AppsV1(),
		BatchV1: k8s.BatchV1(),
	}

	err = gc.ContinuousIntegrationNamespaces(ctx, k8sClients, getDefaultResourceAgeFuncs(), strings.Split(*protectedBranches, ","), strings.Split(*optOutAnnotations, ","), *maxBuildNamespaceAge, *maxReviewNamespaceAge, *dryRun)
	if err != nil {
		log.Fatalf("failed to clean up ci namespaces: %s", err)
	}
}

func provideKubernetesClient(kubeconfig string) (*kubernetes.Clientset, error) {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubernetes configuration: %s", err)
	}
	return kubernetes.NewForConfig(k8sConfig)
}

func getDefaultResourceAgeFuncs() []gc.YoungestResourceAgeFunc {
	return []gc.YoungestResourceAgeFunc{
		gc.NamespaceAge,
		gc.YoungestPodAge,
		gc.YoungestDeploymentAge,
	}
}
