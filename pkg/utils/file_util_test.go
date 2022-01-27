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
	"crypto/md5" //#nosec
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestReadFile(t *testing.T) {
	want := "73a0168848db8fc542e2d2f5d4229de0"
	if cwd, err := os.Getwd(); err != nil {
		t.Errorf("Error  %v", err)
	} else {
		if lines, err := ReadFile(path.Join(cwd, "testdata", "read_file.txt")); err != nil {
			t.Errorf("Error reading file %v", err)
		} else {
			str := ToString(lines)
			actual := fmt.Sprintf("%x", md5.Sum([]byte(str))) //#nosec
			if want != actual {
				t.Errorf("Expected has %s but got %s", want, actual)
			}
		}
	}
}

func TestWriteLinesToFile(t *testing.T) {
	lines := []string{
		"v1.23.1+k3s2",
		"v1.23.1+k3s1",
		"v1.22.5+k3s2",
		"v1.22.5+k3s1",
		"v1.21.8+k3s2",
		"v1.21.8+k3s1",
		"v1.20.14+k3s2",
		"v1.20.14+k3s1",
	}
	want := "73a0168848db8fc542e2d2f5d4229de0"
	if cwd, err := os.Getwd(); err != nil {
		t.Errorf("Error  %v", err)
	} else {
		if err := WriteLinesToFile(path.Join(cwd, "testdata", "write_file.out"), lines); err != nil {
			t.Errorf("Error writing file %v", err)
		} else {
			if b, err := ioutil.ReadFile(path.Join(cwd, "testdata", "write_file.out")); err != nil {
				t.Errorf("Error  reading outfile %v", err)
			} else {
				actual := fmt.Sprintf("%x", md5.Sum(b)) //#nosec
				if want != actual {
					t.Errorf("Expected has %s but got %s", want, actual)
				}
			}
		}
	}
	os.Remove(path.Join(cwd, "testdata", "write_file.out"))
}
