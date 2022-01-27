/*
 * Copyright © 2022  Kamesh Sampath <kamesh.sampath@hotmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package utils

import (
	"github.com/elliotchance/orderedmap"
	"github.com/google/go-cmp/cmp"
	"github.com/kameshsampath/go-kluster/pkg/model"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type kubeUtilsTestCase struct {
	kluster             *model.Kluster
	kubeConfigFile      string
	kubeconfig          string
	clustersCount       int
	clusterNames        []string
	contextsCount       int
	contextNames        []string
	usersCount          int
	userNames           []string
	extensionsCount     int
	prefExtensionsCount int
}

var (
	cwd string
	err error
)

func init() {
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func TestUpdateKubeServerNameAndIP(t *testing.T) {
	kcutil, err := NewKubeConfigUtil(path.Join(cwd, "testdata", "config3.out"))
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	want := struct {
		clusterNames  []string
		kuberServerIP string
	}{
		clusterNames:  []string{"demo3"},
		kuberServerIP: "https://192.168.10.1:6443",
	}
	k := &model.Kluster{
		Name:        "demo3",
		IPAddresses: []string{"192.168.10.1", "10.0.0.1", "10.0.10.20"},
	}
	b, err := os.ReadFile(path.Join(cwd, "testdata", "config3"))

	if err != nil {
		os.Remove(path.Join(cwd, "testdata", "config3.out"))
		t.Fatalf("Error %v", err)
	}

	if err := WriteLinesToFile(path.Join(cwd, "testdata", "config3.out"), strings.Split(string(b), "\n")); err != nil {
		t.Fatalf("Error %v", err)
	}
	err = kcutil.updateKubeClusterNameAndIP(k)
	if err != nil {
		os.Remove(path.Join(cwd, "testdata", "config3.out"))
		t.Fatalf("Error %v", err)
	}
	clusterNames, err := Evaluate(path.Join(cwd, "testdata", "config3.out"), yqClustersNameExpr)

	if err != nil || len(clusterNames) == 0 {
		os.Remove(path.Join(cwd, "testdata", "config3.out"))
		t.Fatalf("Expecting atleast one cluster but  %v", err)
	}

	if !reflect.DeepEqual(want.clusterNames, clusterNames) {
		t.Errorf("Diff want-got \n%s", cmp.Diff(want.clusterNames, clusterNames))
	}

	kuberServerIPs, err := Evaluate(path.Join(cwd, "testdata", "config3.out"), ".clusters.[] | .cluster.server")

	if err != nil || len(kuberServerIPs) == 0 {
		os.Remove(path.Join(cwd, "testdata", "config3.out"))
		t.Fatalf("Expecting atleast one cluster but  %v", err)
	}

	if kuberServerIPs[0] != want.kuberServerIP {
		os.Remove(path.Join(cwd, "testdata", "config3.out"))
		t.Fatalf("Expecting atleast cluster address to be %s but got  %s", want.kuberServerIP, kuberServerIPs[0])
	}

	os.Remove(path.Join(cwd, "testdata", "config3.out"))
}

func TestMergeKubeConfig(t *testing.T) {
	kubeConfigsTest := orderedmap.NewOrderedMap()

	kubeConfigsTest.Set("k1Config", kubeUtilsTestCase{
		kluster: &model.Kluster{
			Name:        "demo1",
			IPAddresses: []string{"192.168.10.1"},
		},
		kubeConfigFile:      path.Join(cwd, "testdata", "config1"),
		kubeconfig:          path.Join(cwd, "testdata", "config.out"),
		clustersCount:       1,
		clusterNames:        []string{"demo1"},
		usersCount:          1,
		userNames:           []string{"demo1"},
		contextsCount:       1,
		contextNames:        []string{"demo1"},
		extensionsCount:     0,
		prefExtensionsCount: 0,
	})
	kubeConfigsTest.Set("k2Config", kubeUtilsTestCase{
		kluster: &model.Kluster{
			Name:        "demo2",
			IPAddresses: []string{"192.168.10.2"},
		},
		kubeConfigFile:      path.Join(cwd, "testdata", "config2"),
		kubeconfig:          path.Join(cwd, "testdata", "config.out"),
		clustersCount:       2,
		clusterNames:        []string{"demo1", "demo2"},
		usersCount:          2,
		userNames:           []string{"demo1", "demo2"},
		contextsCount:       2,
		contextNames:        []string{"demo1", "demo2"},
		extensionsCount:     0,
		prefExtensionsCount: 0,
	})
	kubeConfigsTest.Set("noDuplicates", kubeUtilsTestCase{
		kluster: &model.Kluster{
			Name:        "demo2",
			IPAddresses: []string{"192.168.10.2"},
		},
		kubeConfigFile:      path.Join(cwd, "testdata", "config2"),
		kubeconfig:          path.Join(cwd, "testdata", "config.out"),
		clustersCount:       2,
		clusterNames:        []string{"demo1", "demo2"},
		usersCount:          2,
		userNames:           []string{"demo1", "demo2"},
		contextsCount:       2,
		contextNames:        []string{"demo1", "demo2"},
		extensionsCount:     0,
		prefExtensionsCount: 0,
	})

	for _, key := range kubeConfigsTest.Keys() {
		name := key.(string)
		val, _ := kubeConfigsTest.Get(key)
		tc := val.(kubeUtilsTestCase)
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv("KUBECONFIG", tc.kubeconfig)
			if kcutil, err := NewKubeConfigUtil(tc.kubeconfig); err != nil {
				t.Errorf("Error: %v", err)
			} else {
				if newKubeConfigFileContents, err := ReadFile(tc.kubeConfigFile); err != nil {
					t.Errorf("Error: %#v", err)
				} else if err := kcutil.mergeConfigs(tc.kluster, newKubeConfigFileContents); err != nil {
					t.Errorf("Error: %#v", err)
				} else {
					tc.asserts(t)
				}
			}
		})
	}
	//Clean up
	os.Remove(path.Join(cwd, "testdata", "config.out"))
	os.Remove(path.Join(cwd, "testdata", "config3.out"))
}

func TestRemoveContextFromKubeConfig(t *testing.T) {
	removeConfigTest := map[string]kubeUtilsTestCase{
		"k1Config": {
			kluster: &model.Kluster{
				Name:        "demo1",
				IPAddresses: []string{"192.168.10.1"},
			},
			kubeConfigFile:      path.Join(cwd, "testdata", "config1"),
			kubeconfig:          path.Join(cwd, "testdata", "removeconfig"),
			clustersCount:       1,
			clusterNames:        []string{"demo2"},
			usersCount:          1,
			userNames:           []string{"demo2"},
			contextsCount:       1,
			contextNames:        []string{"demo2"},
			extensionsCount:     0,
			prefExtensionsCount: 0,
		},
	}
	for name, tc := range removeConfigTest {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv("KUBECONFIG", tc.kubeconfig)
			if kcutil, err := NewKubeConfigUtil(tc.kubeconfig); err != nil {
				t.Errorf("Error: %v", err)
			} else {
				if err := kcutil.RemoveEntriesFromKubeConfig(tc.kluster.Name); err != nil {
					t.Errorf("Error: %v", err)
				} else {
					tc.asserts(t)
				}
			}
		})
	}
	//reset test data
	b, err := os.ReadFile(path.Join(cwd, "testdata", "removeconfig.orig"))
	if err == nil {
		f := path.Join(cwd, "testdata", "removeconfig")
		os.Create(f)
		os.WriteFile(f, b, 0600)
	}
}

func (tc kubeUtilsTestCase) asserts(t *testing.T) {
	actuals, err := Evaluate(tc.kubeconfig, yqClustersCountExpr)
	assertCounts(t, "clusters", tc.clustersCount, actuals, err)
	actuals, err = Evaluate(tc.kubeconfig, yqClustersNameExpr)
	assertNames(t, "clusters", tc.clusterNames, actuals, err)

	actuals, err = Evaluate(tc.kubeconfig, yqContextsCountExpr)
	assertCounts(t, "contexts", tc.contextsCount, actuals, err)
	actuals, err = Evaluate(tc.kubeconfig, yqContextsNameExpr)
	assertNames(t, "contexts", tc.contextNames, actuals, err)

	actuals, err = Evaluate(tc.kubeconfig, yqUsersCountExpr)
	assertCounts(t, "users", tc.usersCount, actuals, err)
	actuals, err = Evaluate(tc.kubeconfig, yqUsersNameExpr)
	assertNames(t, "users", tc.userNames, actuals, err)

	actuals, err = Evaluate(tc.kubeconfig, yqExtensionsCountExpr)
	assertCounts(t, "extensions", tc.extensionsCount, actuals, err)

	actuals, err = Evaluate(tc.kubeconfig, yqPrefExtensionsCountExpr)
	assertCounts(t, "preferences.extensions", tc.prefExtensionsCount, actuals, err)
}

func assertCounts(t *testing.T, name string, expected int, actuals []string, err error) {
	if err != nil {
		t.Errorf("%s Error: %#v", t.Name(), err)
	}
	if actual, err := strconv.Atoi(actuals[0]); err != nil {
		t.Errorf("%s Error: %#v", t.Name(), err)
	} else {
		if expected != actual {
			t.Errorf("Expected number of %s is %d but got %d  ", name, expected, actual)
		}
	}
}

func assertNames(t *testing.T, name string, expected, actuals []string, err error) {
	if err != nil {
		t.Errorf("%s Error: %#v", t.Name(), err)
	}
	if !reflect.DeepEqual(expected, actuals) {
		t.Errorf("Expected number of %s is %#v but got %#v  ", name, expected, actuals)
	}
}
