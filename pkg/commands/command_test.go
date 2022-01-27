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
	"fmt"
	"github.com/kameshsampath/go-kluster/pkg/utils"
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
					if k.HasKluster(tc.destroyOptions.profile) != nil {
						t.Error("Error expecting Kluster not to exist but it does")
					}
				}
			}
		})
	}
}
