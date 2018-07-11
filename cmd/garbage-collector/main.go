package main

import (
	"flag"

	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/utopia-planitia/kubernetes-gitlab-garbage-collector/lib"
)

func main() {
	var kubeconfig *string
	var gitlabRunnerNamespace *string
	kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	gitlabRunnerNamespace = flag.String("gitlabRunnerNamespace", "gitlab-runner", "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	client, err := k8sClient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	gc.GitlabExecutors(client.CoreV1().Pods(*gitlabRunnerNamespace))
}
