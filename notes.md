# PoA

- 1 - get filtered namespaces with ContinuousIntegrationNamespaces(...)
    - 1.1 - modify ContinuousIntegrationNamespaces to return the cleaned namespace list
- 2 - loop through namespace list
    - 2.1 - loop through all pods within namespaces
    - 2.2 - check pod age with `age := age(pod.ObjectMeta.CreationTimestamp)`
        - check GitlabExecutors, it seems to already delete pods with age > max_age 
    - 2.3 - if age > max_age, remove from gc
- 3 - command line argument to change namespace/pod age condition?

## Commands:

- Run k8s-gitlab-gc-test with local config:

    `go run cmd/k8s-gitlab-gc-test/main.go --kubeconfig=/home/mariosaleiro/.kube/kind_config`
    - Output:
    ```
    namespaces:
    default
    kube-node-lease
    kube-public
    kube-system
    local-path-storage
    ```
- Run main.go with local config:

    `go run . --kubeconfig=/home/mariosaleiro/.kube/kind_config`
    - Output:
    ```
    2022/07/15 11:30:39 kubeconfig: /home/mariosaleiro/.kube/kind_config
    2022/07/15 11:30:39 gitlabRunnerNamespace: gitlab-runner
    2022/07/15 11:30:39 protectedBranches: develop,master,main,preview,review,stage,staging
    2022/07/15 11:30:39 maxGitlabExecutorAge: 4200
    2022/07/15 11:30:39 maxReviewNamespaceAge: 172800
    2022/07/15 11:30:39 maxBuildNamespaceAge: 7200
    2022/07/15 11:30:39 optOutAnnotations: disable-automatic-garbage-collection
    ```