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

package commands

import (
	"crypto/md5" //#nosec
	"fmt"
	"github.com/kameshsampath/kluster/pkg/utils"
	"os"
	"path"
	"testing"
)

type CommandsTest struct {
	startOptions   StartOptions
	destroyOptions DestroyOptions
	kubeConfigFile string
	release        string
	state          string
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

func TestStartKluster(t *testing.T) {
	startCommandTests := map[string]CommandsTest{
		"defaults": {
			startOptions: StartOptions{
				profile:        "demo1",
				memory:         "4G",
				cpus:           2,
				diskSize:       "20G",
				withKubeConfig: true,
			},
			kubeConfigFile: path.Join(cwd, "testdata", "config.out"),
			release:        "20.04 LTS",
			state:          "Running",
		},
	}
	for name, tc := range startCommandTests {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv("KUBECONFIG", tc.kubeConfigFile)
			rootCmd := NewRootCommand()
			rootCmd.SetArgs([]string{"start", "--profile", tc.startOptions.profile, "-v", "debug"})
			if err := rootCmd.Execute(); err != nil {
				t.Errorf("Error running comand %v", err)
			} else {
				k, err := utils.KlusterDetails(tc.startOptions.profile)
				if k != nil {
					info := k.Info
					if info.Name != tc.startOptions.profile {
						t.Errorf("Expected number of %s is %s", tc.startOptions.profile, k.Info.Name)
					}
					if len(info.IPAddresses) == 0 {
						t.Errorf("Expecting the Kluster %s to have atleast one IP but got %d ", tc.startOptions.profile, 0)
					}
					if info.Release != "20.04 LTS" {
						t.Errorf("Expected release to be %s but it is %s", tc.release, info.Release)
					}
					if info.State != "Running" {
						t.Errorf("Expected state to be %s but it is %s", tc.state, info.State)
					}
				} else {
					t.Errorf("Error expecting Kluster to exist but %#v", err)
				}
			}
		})
		deleteOpts := DestroyOptions{
			profile: tc.startOptions.profile,
		}
		if _, err := deleteOpts.destroyKluster(); err != nil {
			t.Logf("Unable to cleanup the kluster %s. Delete it manually", tc.startOptions.profile)
		}
	}
}

func TestDestroyKluster(t *testing.T) {
	startOpts := StartOptions{
		profile:        "demo1",
		memory:         "4G",
		cpus:           2,
		diskSize:       "20G",
		withKubeConfig: true,
	}

	if _, err := startOpts.startKluster(); err != nil {
		t.Fatalf("Error %v", err)
	}

	destroyCommandTests := map[string]CommandsTest{
		"defaults": {
			destroyOptions: DestroyOptions{
				profile:           "demo1",
				removeKubeContext: false,
			},
			kubeConfigFile: path.Join(cwd, "testdata", "removeconfig"),
		},
		"removeKubeContext": {
			destroyOptions: DestroyOptions{
				profile:           "demo1",
				removeKubeContext: true,
			},
			kubeConfigFile: path.Join(cwd, "testdata", "removeconfig"),
		},
	}
	for name, tc := range destroyCommandTests {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv("KUBECONFIG", tc.kubeConfigFile)
			rootCmd := NewRootCommand()
			rootCmd.SetArgs([]string{"destroy", "--profile", tc.destroyOptions.profile, "-v", "debug"})
			if err := rootCmd.Execute(); err != nil {
				kerr := fmt.Sprintf("kluster \"%s\" does not exist", tc.destroyOptions.profile)
				if err.Error() != kerr {
					t.Fatalf("Error running comand %v", err)
				}
			} else {
				k, _ := utils.Klusters()
				if k != nil {
					if k.GetKluster(tc.destroyOptions.profile) != nil {
						t.Error("Error expecting Kluster not to exist but it does")
					}
				}
			}
		})
	}
	deleteOpts := DestroyOptions{
		profile: startOpts.profile,
	}
	if _, err := deleteOpts.destroyKluster(); err != nil {
		t.Logf("Unable to cleanup the kluster %s. Delete it manually", startOpts.profile)
	}
}

func TestKubeConfig(t *testing.T) {
	startOpts := StartOptions{
		profile:        "demo4",
		memory:         "4G",
		cpus:           2,
		diskSize:       "20G",
		withKubeConfig: false,
	}

	if _, err := startOpts.startKluster(); err != nil {
		t.Fatalf("Error %v", err)
	}

	kubeconfigTestCases := map[string]struct {
		kubeConfigOptions KubeConfigOptions
		kubeConfigFile    string
		want              string
	}{
		"defaults": {
			kubeConfigOptions: KubeConfigOptions{
				profile: "demo4",
			},
			want:           "b18e75feb4997840fca96effd51ba023",
			kubeConfigFile: path.Join(cwd, "testdata", "config.out"),
		},
		"customPath": {
			kubeConfigOptions: KubeConfigOptions{
				profile:        "demo4",
				kubeConfigFile: path.Join(cwd, "testdata", "custom.out"),
			},
			want:           "b18e75feb4997840fca96effd51ba023",
			kubeConfigFile: path.Join(cwd, "testdata", "custom.out"),
		},
	}

	for name, tc := range kubeconfigTestCases {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv("KUBECONFIG", tc.kubeConfigFile)
			rootCmd := NewRootCommand()
			if tc.kubeConfigOptions.kubeConfigFile != "" {
				rootCmd.SetArgs([]string{"kubeconfig", "--profile", tc.kubeConfigOptions.profile, "--to-file", tc.kubeConfigOptions.kubeConfigFile, "-v", "debug"})
			} else {
				rootCmd.SetArgs([]string{"kubeconfig", "--profile", tc.kubeConfigOptions.profile, "-v", "debug"})
			}

			if err := rootCmd.Execute(); err != nil {
				b, err := os.ReadFile(tc.kubeConfigFile)
				if err != nil {
					t.Fatalf("Error %v", err)
				}
				actual := fmt.Sprintf("%x", md5.Sum(b)) //#nosec
				if tc.want != actual {
					t.Errorf("Expecting checksum to be %s but got %s", tc.want, actual)
				}
			}
		})
	}
	os.Remove(path.Join(cwd, "testdata", "config.out"))
	os.Remove(path.Join(cwd, "testdata", "custom.out"))
	deleteOpts := DestroyOptions{
		profile: startOpts.profile,
	}
	if _, err := deleteOpts.destroyKluster(); err != nil {
		t.Logf("Unable to cleanup the kluster %s. Delete it manually", startOpts.profile)
	}
}
