package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

/*
PoA:
  - Setup
    - k8s-gitlab-gc deploy in cluster in namespace
		- configure ci namespace deletion with 15s
		- configure review namespace deletion with 2d
	- wait for k8s-gitlab-gc deployment
	- ✅ create 'review' namespace
	- ✅ create 'ci-test-1' namespace
	- ✅ wait 15s
	- ✅ create 'ci-test-2' namespace
		- (optional) add pod (decide timing later)
  - Assess
	- wait for another 2 seconds
	- check 'test-1' namespace was
	- check 'test-2' namespace still exists
	- check 'review' namepsace still exists
  - Repeat until we figure out the minimum deletion time for testing
*/

/*
Open questions:
  - Since we always make the resource creation after creating a resource/namespace/pod,
	should we pass the context into the createNamespace/createPod and make the resource
	creation inside the function? e.g.:

	func newNamespace(ctx context.Context, name string) *corev1.Namespace {
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}
		if err := cfg.Client().Resources().Create(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		return namespace

	}
*/

var gcJobName string = "k8s-gitlab-gc"
var ciTestNamespaceOne string = "ci-test-1-d41d8cd98f00b204e9800998ecf8427e"
var ciTestNamespaceTwo string = "ci-test-2-d41d8cd98f00b204e9800998ecf8427e"

func TestRealCluster(t *testing.T) {

	deploymentFeature := features.New("batchv1/job").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {

			t.Log("wait for default service account")
			sAccounts := &corev1.ServiceAccountList{
				Items: []corev1.ServiceAccount{
					*defaultServiceAccount(cfg.Namespace()),
				},
			}
			err := wait.For(conditions.New(cfg.Client().Resources()).ResourcesFound(sAccounts))
			if err != nil {
				t.Fatal(err)
			}

			// give default serviceaccount cluster admin role in this namespace
			// is used by the k8s-gitlab-gc for the deletion of gitlab-runner pods
			t.Log("create ClusterRoleBinding for default service account")
			crb := newClusterRoleBinding("gitlab-ci-runner", cfg.Namespace())
			if err := cfg.Client().Resources().Create(ctx, crb); err != nil {
				t.Fatal(err)
			}

			t.Log("create review namespace")
			reviewNamespace := newNamespace("review")
			if err := cfg.Client().Resources().Create(ctx, reviewNamespace); err != nil {
				t.Fatal(err)
			}

			t.Log("create 1st ci namespace")
			ciTestNamespace1 := newNamespace(ciTestNamespaceOne)
			if err := cfg.Client().Resources().Create(ctx, ciTestNamespace1); err != nil {
				t.Fatal(err)
			}

			t.Log("create 2nd ci namespace")
			ciTestNamespace2 := newNamespace(ciTestNamespaceTwo)
			if err := cfg.Client().Resources().Create(ctx, ciTestNamespace2); err != nil {
				t.Fatal(err)
			}

			namespaces := &corev1.NamespaceList{
				Items: []corev1.Namespace{
					*reviewNamespace,
					*ciTestNamespace1,
					*ciTestNamespace2,
				},
			}
			t.Log("wait for namespaces")
			err = wait.For(conditions.New(cfg.Client().Resources()).ResourcesFound(namespaces))
			if err != nil {
				t.Fatal(err)
			}

			t.Log("sleep 15s")
			time.Sleep(15 * time.Second)

			t.Log("create pod in 2nd ci namespace")
			testPod := newPod("nginx-pod", ciTestNamespace2)
			if err := cfg.Client().Resources().Create(context.TODO(), testPod); err != nil {
				t.Error("failed to create pod due to an error", err)
			}

			// t.Log("sleep 10s")
			// time.Sleep(10 * time.Second)

			return ctx
		}).
		Assess("namespace deletion", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			gcImageName := fmt.Sprintf("%v", ctx.Value(keyImageName))

			t.Log("create gc job")
			gcJob := newGcJob(cfg.Namespace(), gcJobName, gcImageName)
			if err := cfg.Client().Resources().Create(ctx, gcJob); err != nil {
				t.Fatal(err)
			}

			t.Log("start - wait for gc job")
			err := wait.For(conditions.New(cfg.Client().Resources()).JobCompleted(gcJob),
				wait.WithTimeout(time.Minute*1),
			)
			if err != nil {
				logs, err2 := getJobLogs(ctx, gcJob, cfg.Client().RESTConfig())
				if err2 != nil {
					t.Logf("job container image: %s", gcJob.Spec.Template.Spec.Containers[0].Image)
					t.Logf("error - failed to wait for job condition: %s", err)
					t.Fatalf("unable to get logs: %s", err2)
				}

				t.Logf("job logs:\n %s\n", logs)

				t.Fatalf("failed to wait for job condition: %s", err)
			}
			t.Log("finish - wait for gc job")

			t.Log("check 2nd ci namespace was not deleted")
			ciNsTwo := newNamespace(ciTestNamespaceTwo)
			// err = wait.For(conditions.New(cfg.Client().Resources()).ResourcesFound(namespaces), wait.WithTimeout(time.Second*1))
			err = wait.For(conditions.New(cfg.Client().Resources()).ResourceMatch(ciNsTwo, func(object k8s.Object) bool {
				n := object.(*corev1.Namespace)
				t.Logf("namespace phase of namespace \"%s\" is \"%s\"", n.Name, n.Status.Phase)
				return n.Status.Phase != corev1.NamespaceTerminating
			}), wait.WithTimeout(time.Second*1))
			if err != nil {
				t.Fatalf("error - 2nd ci namespace not found or is terminating: %s", err)
			}

			t.Log("check 1st ci namespace was deleted")
			err = wait.For(conditions.New(cfg.Client().Resources()).ResourceDeleted(newNamespace(ciTestNamespaceOne)), wait.WithTimeout(time.Second*30))
			if err != nil {
				logs, err2 := getJobLogs(ctx, gcJob, cfg.Client().RESTConfig())
				if err2 != nil {
					t.Logf("job container image: %s", gcJob.Spec.Template.Spec.Containers[0].Image)
					t.Logf("error - 1st ci namespace was not deleted: %s", err)
					t.Fatalf("unable to get logs: %s", err2)
				}

				t.Logf("job logs:\n %s\n", logs)

				t.Fatalf("error - 1st ci namespace was not deleted: %s", err)
			}

			// t.Log("sleep 5s")
			// time.Sleep(5 * time.Second)

			return context.WithValue(ctx, keyGcJob, gcJob)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var job *batchv1.Job = ctx.Value(keyGcJob).(*batchv1.Job)
			if err := cfg.Client().Resources().Delete(ctx, job); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testenv.Test(t, deploymentFeature)
}

func newGcJob(namespace string, name string, image string) *batchv1.Job {
	// backoffLimit=0 - only run k8s job once, even when it should fail
	backoffLimit := int32(0)
	restartPolicy := corev1.RestartPolicyNever

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: namespace, Labels: map[string]string{"job": name},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"job": name}},
				Spec: corev1.PodSpec{
					RestartPolicy: restartPolicy,
					Containers: []corev1.Container{
						{
							// note: ImagePullPolicy is needed if one uses latest tag
							// see: https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            "gc", Image: image,
							Command: []string{"/k8s-gitlab-gc"},
							Args: []string{
								"-maxReviewNamespaceAge=300",
								"-maxBuildNamespaceAge=15",
								fmt.Sprintf("-gitlabRunnerNamespace=%s", namespace),
							},
						},
					},
				},
			},
		},
	}
}

func getJobLogs(ctx context.Context, job *batchv1.Job, restCfg *rest.Config) (string, error) {
	podLogOpts := corev1.PodLogOptions{}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return "", fmt.Errorf("error in getting access to K8S: %s", err)
	}

	pods, err := clientset.CoreV1().Pods(job.Namespace).List(
		context.TODO(),
		metav1.ListOptions{LabelSelector: fmt.Sprintf("job=%s", job.Name)},
	)
	if err != nil {
		return "", fmt.Errorf("error in list of pods: %s", err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("%s", "pod list is empty")
	}

	jobPod := pods.Items[0]
	log.Printf("job pod contaner image: \"%s\"", jobPod.Spec.Containers[0].Image)

	req := clientset.CoreV1().Pods(job.Namespace).GetLogs(jobPod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error in opening stream: %s", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("error in copy information from podLogs to buf: %s", err)
	}
	str := buf.String()

	return str, nil
}

func newClusterRoleBinding(name string, namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func defaultServiceAccount(namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}
}

func newNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
}

func newPod(name string, namespace *corev1.Namespace) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace.ObjectMeta.Name,
			Labels:    map[string]string{"app": name},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name,
					Image: "nginx",
				},
			},
		},
	}
}
