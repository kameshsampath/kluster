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

package commands

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kameshsampath/kluster/pkg/utils"
	"os"
	"path"
	"strings"
	"testing"
)

func init() {
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func TestCloudInitFileGeneration(t *testing.T) {
	k3sUtil, err := utils.NewK3sVersionUtil(path.Join(cwd, "testdata", "k3s-releases.out"), "")
	if err != nil {
		panic(err)
	}
	startOpts := StartOptions{
		profile: "demo1",
	}
	if c, err := NewCloudInitUtil(k3sUtil); err != nil {
		t.Errorf("Error Getting Cloud Init %v", err)
	} else {
		if err := c.configureAndGetCloudInitFile(&startOpts); err != nil {
			t.Errorf("Error Configuring Cloud Init %v", err)
		} else {
			_, err := os.Stat(startOpts.cloudInitFile)
			if err != nil {
				t.Errorf("Expecting cloud Init file %s to exist %v", startOpts.cloudInitFile, err)
			} else {
				if !strings.HasPrefix(path.Base(startOpts.cloudInitFile), startOpts.profile) {
					t.Errorf("Expecting cloud Init file to have prefix %s but it is %s", startOpts.profile, path.Base(startOpts.cloudInitFile))
				}
			}
		}
	}
	os.Remove(startOpts.cloudInitFile)
}
func TestCloudInitK3sCommands(t *testing.T) {
	k3sUtil, err := utils.NewK3sVersionUtil(path.Join(cwd, "testdata", "k3s-releases.out"), "")
	if err != nil {
		panic(err)
	}
	c, err := NewCloudInitUtil(k3sUtil)
	if err != nil {
		panic(err)
	}
	k3sCloudInitTestCases := map[string]struct {
		startOpts StartOptions
		want      []string
	}{
		"defaults": {
			startOpts: StartOptions{
				profile: "demo1",
			},
			want: []string{
				fmt.Sprintf(k3sInstallCmd, k3sUtil.Versions[0], ""),
				k3sKubeConfigCopyCmd,
			},
		},
		"customK3sVersion": {
			startOpts: StartOptions{
				profile:    "demo1",
				k3sVersion: "v1.21.8+k3s1",
			},
			want: []string{
				fmt.Sprintf(k3sInstallCmd, "v1.21.8+k3s1", ""),
				k3sKubeConfigCopyCmd,
			},
		},
		"noTrafeik": {
			startOpts: StartOptions{
				profile:        "demo1",
				k3sVersion:     "v1.21.8+k3s1",
				k3sServerFlags: []string{"--disable traefik"},
			},
			want: []string{
				fmt.Sprintf(k3sInstallCmd, "v1.21.8+k3s1", "--disable traefik"),
				k3sKubeConfigCopyCmd,
			},
		},
	}

	for name, tc := range k3sCloudInitTestCases {
		startOpts := tc.startOpts
		t.Run(name, func(t *testing.T) {
			if err := c.configureAndGetCloudInitFile(&startOpts); err != nil {
				t.Errorf("Error Configuring Cloud Init %v", err)
			} else {
				_, err := os.Stat(startOpts.cloudInitFile)
				if err != nil {
					t.Errorf("Expecting cloud Init file %s to exist %v", startOpts.cloudInitFile, err)
				} else {
					if !strings.HasPrefix(path.Base(startOpts.cloudInitFile), startOpts.profile) {
						t.Errorf("Expecting cloud Init file to have prefix %s but it is %s", startOpts.profile, path.Base(startOpts.cloudInitFile))
					}
				}
				l := len(c.Runcmd)
				if tc.want[0] != c.Runcmd[l-2] {
					t.Errorf("Diff want - got \n%s", cmp.Diff(tc.want[0], c.Runcmd[l-2]))
				}
				if tc.want[1] != c.Runcmd[l-1] {
					t.Errorf("Diff want - got \n%s", cmp.Diff(tc.want[1], c.Runcmd[l-1]))
				}
			}
		})
	}
	os.Remove(path.Join(cwd, "testdata", "k3s-releases"))
}
