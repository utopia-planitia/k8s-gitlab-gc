package gc

// import (
// 	"os"
// 	"testing"

// 	"sigs.k8s.io/e2e-framework/pkg/env"
// 	"sigs.k8s.io/e2e-framework/pkg/envconf"
// 	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
// )

// var (
// 	testenv env.Environment
// )

// func TestMain(m *testing.M) {
// 	testenv = env.New()
// 	kindClusterName := envconf.RandomName("e2e-cluster", 16)
// 	namespace := envconf.RandomName("e2ens", 16)

// 	// Use pre-defined environment funcs to create a kind cluster prior to test run
// 	testenv.Setup(
// 		envfuncs.CreateKindCluster(kindClusterName),
// 		envfuncs.CreateNamespace(namespace),
// 	)

// 	// Use pre-defined environment funcs to teardown kind cluster after tests
// 	testenv.Finish(
// 		envfuncs.DeleteNamespace(namespace),
// 		envfuncs.DestroyKindCluster(kindClusterName),
// 	)

// 	//launch package tests
// 	os.Exit(testenv.Run(m))
// }

// // func TestKubernetes(t *testing.T) {
// // 	f1 := features.New("count pod").
// // 		WithLabel("type", "pod-count").
// // 		Assess("pods from kube-system", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
// // 			var pods corev1.PodList
// // 			err := cfg.Client().Resources("kube-system").List(context.TODO(), &pods)
// // 			if err != nil {
// // 				t.Fatal(err)
// // 			}
// // 			if len(pods.Items) == 0 {
// // 				t.Fatal("no pods in namespace kube-system")
// // 			}
// // 			return ctx
// // 		}).Feature()

// // 	f2 := features.New("count namespaces").
// // 		WithLabel("type", "ns-count").
// // 		Assess("namespace exist", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
// // 			var nspaces corev1.NamespaceList
// // 			err := cfg.Client().Resources().List(context.TODO(), &nspaces)
// // 			if err != nil {
// // 				t.Fatal(err)
// // 			}
// // 			if len(nspaces.Items) == 1 {
// // 				t.Fatal("no other namespace")
// // 			}
// // 			return ctx
// // 		}).Feature()

// // 	// test feature
// // 	testenv.Test(t, f1, f2)
// // }
