package main

import (
	"flag"

	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/utopia-planitia/kubernetes-gitlab-garbage-collector/lib"
)

func main() {

	// TODO: remove gitlab runners if older then 1h
	// TODO: remove ci namespaces if nothing got updated for 2 days (only clean up .*-ci-.* and keep master / stage / develop)
	// TODO: remove ci namespaces if branch is gone
	// TODO: remove gitlab environments if ingress is gone

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

	gc.RunnerPods(client.CoreV1().Pods(*gitlabRunnerNamespace))
}
