// https://github.com/kubernetes-sigs/e2e-framework/blob/main/examples/real_cluster/main_test.go
package e2e

import (
	// If testing with a cloud vendor managed cluster uncomment one of the below dependencies to properly get authorised.
	//_ "k8s.io/client-go/plugin/pkg/client/auth/azure" // auth for AKS clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"   // auth for GKE clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"  // auth for OIDC

	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

type key int

const (
	keyImageName key = iota
	keyGcJob
)

var testenv env.Environment

func TestMain(m *testing.M) {
	//create kind if doesn't already exist

	testenv = env.New()
	namespace := envconf.RandomName("k8s-gitlab-gc", 16)
	if os.Getenv("REAL_CLUSTER") == "true" {
		// path := conf.ResolveKubeConfigFile()
		// cfg := envconf.NewWithKubeConfig(path)
		// testenv = env.NewWithConfig(cfg)

		// testenv.Setup(
		// 	envfuncs.CreateNamespace(namespace),
		// )
		// testenv.Finish(
		// 	envfuncs.DeleteNamespace(namespace),
		// )
		log.Fatal("not implemented yet.")
	} else {
		// get current working dir, to resolve the path to Dockerfile
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		buildContextPath := filepath.Join(wd, "../..")
		fmt.Println("Build Context Path: ", buildContextPath)

		// form a unique docker gcImageNameAndTag. the first string seg is the local docker registry host
		//gcImageNameAndTag := fmt.Sprintf("%s%s%s", "docker-registry:5000/", build.Name(), p.md5()[:6])
		gcImageNameAndTag := "k8s-gitlab-gc"
		// gcImageNameAndTag := fmt.Sprintf("%s%s%s", "docker-registry:5000/", "k8s-gitlab-gc", ":latest")

		_, err = dockerBuild(gcImageNameAndTag, buildContextPath)
		if err != nil {
			log.Fatal(err)
		}

		kindClusterName := envconf.RandomName("k8s-gitlab-gc-kind", 25)

		testenv.Setup(
			envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
			// note: if you use latest tage you need to use ImagePullPolicy `IfNotPresented` or `Never``
			//   see: https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster
			envfuncs.LoadDockerImageToCluster(kindClusterName, gcImageNameAndTag),
			envfuncs.CreateNamespace(namespace),
			func(ctx context.Context, env *envconf.Config) (context.Context, error) {
				return context.WithValue(ctx, keyImageName, gcImageNameAndTag), nil
			},
		)

		testenv.Finish(
			envfuncs.DeleteNamespace(namespace),
			envfuncs.DestroyCluster(kindClusterName),
		)
	}

	os.Exit(testenv.Run(m))
}

func dockerBuild(imageTag string, buildContextPath string) (buildOut []byte, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// create a docker buildContext by `archiving` the files
	// the target dir
	buildCtx, err := createBuildContext(buildContextPath)
	if err != nil {
		return []byte{}, err
	}

	start := time.Now()

	ctx := context.Background()
	// build image. reader can be used to get output from docker deamon
	reader, err := cli.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile: "Dockerfile", PullParent: true, Tags: []string{imageTag}, Remove: true, NoCache: false,
	})
	if err != nil {
		return []byte{}, err
	}

	scanner := bufio.NewScanner(reader.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		fmt.Println(line)
	}
	// for {
	// 	buff := make([]byte, 512)
	// 	_, err := reader.Body.Read(buff)

	// 	if err != nil {
	// 		if errors.Is(err, io.EOF) {
	// 			break
	// 		}
	// 		return []byte{}, err
	// 	}
	// 	fmt.Println(string(buff[:]))
	// }

	// contents, err := ioutil.ReadAll(reader.Body)
	// if err != nil {
	// 	return []byte{}, err
	// }

	elapsed := time.Since(start)
	log.Printf("Build took %s", elapsed)

	return []byte{}, nil
}

// createBuildContext archive a dir and return an io.Reader
func createBuildContext(path string) (io.Reader, error) {
	//return archive.Tar(path, archive.Uncompressed)
	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: []string{".git", "test/e2e"},
		IncludeFiles:    []string{"."},
	}
	return archive.TarWithOptions(path, options)
}
