/*
 * Copyright (c) 2022.
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *             http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 */
package utils

import (
	"github.com/google/go-cmp/cmp"
	"github.com/kameshsampath/go-kluster/pkg/model"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
)

func init() {
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func TestKlusters(t *testing.T) {
	klustersListTestCases := map[string]struct {
		listJSONFile string
		want         model.KlusterList
	}{
		"defaults": {
			listJSONFile: path.Join(cwd, "testdata", "klusters_list.json"),
			want: model.KlusterList{
				Klusters: []model.Kluster{{
					Name:        "vm1",
					IPAddresses: []string{"192.168.64.92"},
					Release:     "20.04 LTS",
					State:       "Running",
					Memory:      model.Memory{},
				}, {
					Name:        "cluster1",
					IPAddresses: []string{},
					Release:     "20.04 LTS",
					State:       "Stopped",
					Memory:      model.Memory{},
				},
				},
			},
		},
		"noklusters": {
			listJSONFile: path.Join(cwd, "testdata", "no_klusters.json"),
			want:         model.KlusterList{Klusters: nil},
		},
	}

	for name, tc := range klustersListTestCases {
		t.Run(name, func(t *testing.T) {
			if b, err := ioutil.ReadFile(tc.listJSONFile); err != nil {
				t.Errorf("Error loading test JSON %s, %v", tc.listJSONFile, err)
			} else {
				var actual model.KlusterList
				if err := actual.MachineListToKlusterList(b); err != nil {
					t.Errorf("Error getting machine list %v", err)
				} else {
					if tc.want.Len() != actual.Len() {
						t.Errorf("Expecting %d Klusters but got %d", tc.want.Len(), actual.Len())
					}
					sort.Sort(&tc.want)
					sort.Sort(&actual)
					for i := 0; i < tc.want.Len(); i++ {
						if !reflect.DeepEqual(tc.want.Klusters[i], actual.Klusters[i]) {
							t.Errorf("Diff of want - got\n%s", cmp.Diff(tc.want.Klusters[i], actual.Klusters[i]))
						}
					}
				}
			}
		})
	}
}
func TestKluster(t *testing.T) {
	klusterTestCases := map[string]struct {
		listJSONFile string
		want         model.KlusterDetails
	}{
		"cluster1": {
			listJSONFile: path.Join(cwd, "testdata", "cluster1.json"),
			want: model.KlusterDetails{
				Info: model.KlusterInfo{
					Name:        "cluster1",
					IPAddresses: nil,
					Release:     "20.04 LTS",
					State:       "Stopped",
					Memory:      model.Memory{},
				},
				Errors: []string{},
			},
		},
		"vm1": {
			listJSONFile: path.Join(cwd, "testdata", "vm1.json"),
			want: model.KlusterDetails{
				Info: model.KlusterInfo{
					Name:        "vm1",
					IPAddresses: []string{"192.168.64.92"},
					Release:     "20.04 LTS",
					State:       "Running",
					Memory: model.Memory{
						Total: 4122849280,
						Used:  150953984,
					},
				},
				Errors: []string{},
			},
		},
	}

	for name, tc := range klusterTestCases {
		t.Run(name, func(t *testing.T) {
			if b, err := ioutil.ReadFile(tc.listJSONFile); err != nil {
				t.Errorf("Error loading test JSON %s, %v", tc.listJSONFile, err)
			} else {
				var actual model.KlusterDetails
				if err := actual.MachineInfoToKlusterDetails(b); err != nil {
					t.Errorf("Error getting machine list %v", err)
				} else {
					if !reflect.DeepEqual(tc.want.Info, actual.Info) {
						t.Errorf("Diff of want - got\n%s", cmp.Diff(tc.want.Info, actual.Info))
					}
					sort.Strings(tc.want.Errors)
					sort.Strings(actual.Errors)
					if !reflect.DeepEqual(tc.want.Errors, actual.Errors) {
						t.Errorf("Diff of want - got\n%s", cmp.Diff(tc.want.Info, actual.Info))
					}
				}
			}
		})
	}
}
