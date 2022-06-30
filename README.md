# kubernetes & gitlab garbage collector

* why this tool:
  * deletes k8s namespaces
    * we have ci namespaces normally created for the lifetime of a ci pipeline
    * normally when the pipeline fishes successfully the namespace is deleted by the pipline
    * if the pipelin fails the namespace is not deleted
      * that is ok for a certain amount of time as developer might what to debug why the pipeline failed
      * but the namespace also can't stay forever and sometimes we humans just forget to delete namespaces after debuggging or we even not looked them to debug in the firstplace
      * this why we want to give ci namespaces kind of a max lifetime and then be garbe-collected
        * e.g. ci-test namespaces only stay for 2h
        * review environments stay for 2 days
    * it's mainly about these ci namespaces but also any other namespace could be wanted to be garbage-collected
  * deletes gitlab-runner pods
    * when: pod is older than 90m
    * reason:
      * for some reason it can happen that pods belonging to a job exits where the job runned in comletion (fisnihed or failed) but the pod still hangs around
      * bec of that we want to delete idle pods to free resources again
