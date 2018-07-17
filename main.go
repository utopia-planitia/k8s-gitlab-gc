package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/plouc/go-gitlab-client/gitlab"
	"github.com/utopia-planitia/k8s-gitlab-gc/lib"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	optOutAnnotations = flag.String("optOutAnnotations", "disable-automatic-garbage-collection", "comma seperated list of annotations to protect namespaces from deletion, annotations need to be set to the string 'true'")
	var gitlabURL *string
	gitlabURL = flag.String("gitlabURL", "", "url of gitlab")
	var gitlabTokenFile *string
	gitlabTokenFile = flag.String("gitlabTokenFile", "/gitlab-secret/token", "path to the gitlab api token")
	flag.Parse()

	k8s, err := provideKubernetesClient(*kubeconfig)
	if err != nil {
		log.Fatalf("failed initilize kubernetes client: %s", err)
	}
	g, err := provideGitlabClient(*gitlabTokenFile, *gitlabURL)
	if err != nil {
		log.Fatalf("failed initilize gitlab client: %s", err)
	}

	err = gc.GitlabExecutors(k8s.CoreV1().Pods(*gitlabRunnerNamespace), *maxGitlabExecutorAge)
	if err != nil {
		log.Fatalf("failed to clean up gitlab executors: %s", err)
	}

	err = gc.ContinuousIntegrationNamespaces(k8s.CoreV1(), strings.Split(*protectedBranches, ","), strings.Split(*optOutAnnotations, ","), *maxBuildNamespaceAge, *maxReviewNamespaceAge)
	if err != nil {
		log.Fatalf("failed to clean up ci namespaces: %s", err)
	}

	err = gc.GitlabEnvironments(k8s.CoreV1().Nodes(), k8s.ExtensionsV1beta1().Ingresses(""), g)
	if err != nil {
		log.Fatalf("failed to clean up gitlab project environments: %s", err)
	}
}

func provideKubernetesClient(kubeconfig string) (*kubernetes.Clientset, error) {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubernetes configuration: %s", err)
	}
	return kubernetes.NewForConfig(k8sConfig)
}

func provideGitlabClient(tokenPath, url string) (*gitlab.Gitlab, error) {
	b, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read gitlab api token: %s", err)
	}
	gitlabToken := strings.TrimSpace(string(b))
	return gitlab.NewGitlab(url, "/api/v4/", gitlabToken), nil
}
