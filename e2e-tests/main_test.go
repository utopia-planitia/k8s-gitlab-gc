/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gc

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	kindClusterName := envconf.RandomName("kind-with-config", 16)
	namespace := envconf.RandomName("kind-ns", 16)
	fmt.Println("TestMain: starting testenv")

	testenv.Setup(
		envfuncs.CreateKindClusterWithConfig(kindClusterName, "kindest/node:v1.22.2", "kind-config.yaml"),
		envfuncs.CreateNamespace(namespace),
	)
	fmt.Println("TestMain: testenv setup done")

	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
		//envfuncs.DestroyKindCluster(kindClusterName),
	)

	fmt.Println("TestMain: testenv finish done")

	os.Exit(testenv.Run(m))
	fmt.Println("TestMain: done")

}
