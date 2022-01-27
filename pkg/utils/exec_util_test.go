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
	"bytes"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"reflect"
	"testing"
)

func TestExecUtil(t *testing.T) {
	var versionStdOutBuf bytes.Buffer
	versionStdOutBuf.WriteString(fmt.Sprintf("%s\n", "multipass   1.8.1+mac"))
	versionStdOutBuf.WriteString(fmt.Sprintf("%s\n", "multipassd  1.8.1+mac"))

	execTestCases := map[string]struct {
		commandArgs []string
		want        ExecResponse
	}{
		"success": {
			commandArgs: []string{"version"},
			want: ExecResponse{
				Command:        "/usr/local/bin/multipass version",
				StdErr:         bytes.Buffer{},
				StdOut:         versionStdOutBuf,
				StdOutErrLines: sanitize(versionStdOutBuf),
			},
		},
	}
	for name, tc := range execTestCases {
		t.Run(name, func(t *testing.T) {
			actual := NewExecUtil()
			if err := actual.Execute(tc.commandArgs); err != nil {
				t.Errorf("Error %v", err)
			} else {
				if tc.want.Command != actual.Command {
					t.Errorf("Expected %s got %s", tc.want.Command, actual.Command)
				}
				if tc.want.StdErr.String() != actual.StdErr.String() {
					t.Errorf("Expected %s got %s", tc.want.StdErr.String(), actual.StdErr.String())
				}
				if tc.want.StdOut.String() != actual.StdOut.String() {
					t.Errorf("Diff of want - got\n%s", cmp.Diff(tc.want.StdOut.String(), actual.StdOut.String()))
				}
				if !reflect.DeepEqual(tc.want.StdOutErrLines, actual.StdOutErrLines) {
					t.Errorf("Diff of want - got\n%s", cmp.Diff(tc.want.StdOutErrLines, actual.StdOutErrLines))
				}
			}
		})
	}
}
