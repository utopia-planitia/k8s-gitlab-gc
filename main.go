package main

import (
	"flag"
	"log"
	"strings"

	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/utopia-planitia/k8s-gitlab-gc/lib"
)

func main() {
	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	var gitlabRunnerNamespace *string
	gitlabRunnerNamespace = flag.String("gitlabRunnerNamespace", "gitlab-runner", "namespace to remove gitlab executors from")
	var protectedBranches *string
	protectedBranches = flag.String("protectedBranches", "develop,master,preview,review,stage,staging", "comma seperated list of substrings to mark a namespace as protected from deletion")
	var maxGitlabExecutorAge *int64
	maxGitlabExecutorAge = flag.Int64("maxGitlabExecutorAge", 60*60*2, "max age for gitlab executor pods in seconds")
	var maxReviewNamespaceAge *int64
	maxReviewNamespaceAge = flag.Int64("maxReviewNamespaceAge", 60*60*24*2, "max age for review namespaces in seconds")
	var maxBuildNamespaceAge *int64
	maxBuildNamespaceAge = flag.Int64("maxBuildNamespaceAge", 60*60*6, "max age for e2e testing namespaces in seconds")
	var optOutAnnotations *string
	optOutAnnotations = flag.String("optOutAnnotations", "disable-automatic-garbage-collection", "comma seperated list of annotations to protect namespaces from deletion, annotations need to be set to the string true")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	client, err := k8sClient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	err = gc.GitlabExecutors(client.CoreV1().Pods(*gitlabRunnerNamespace), *maxGitlabExecutorAge)
	if err != nil {
		log.Printf("failed to clean up gitlab executors: %s", err)
	}
	err = gc.ContinuousIntegrationNamespaces(client.CoreV1(), strings.Split(*protectedBranches, ","), strings.Split(*optOutAnnotations, ","), *maxBuildNamespaceAge, *maxReviewNamespaceAge)
	if err != nil {
		log.Printf("failed to clean up ci namespaces: %s", err)
	}
}
