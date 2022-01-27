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
	"github.com/elliotchance/orderedmap"
	"os"
	"path"
	"testing"
	"time"
)

func init() {
	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

type k3sVersionsTestCase struct {
	cacheFile  string
	expiryTime string
	want       K3sVersionInfo
}

func TestK3sVersions(t *testing.T) {
	// creating before test start to adjust for 1s slack
	os.Create(path.Join(cwd, "testdata", "k3s-releases-expired"))
	k3sVersionsTestCases := orderedmap.NewOrderedMap()
	k3sVersionsTestCases.Set("nocache", k3sVersionsTestCase{
		cacheFile: path.Join(cwd, "testdata", "k3s-releases"),
		want: K3sVersionInfo{
			CacheFile:  path.Join(cwd, "testdata", "k3s-releases"),
			Versions:   []string{},
			ExpiryTime: time.Hour * 24,
			FromCache:  false,
		},
	})
	k3sVersionsTestCases.Set("cached", k3sVersionsTestCase{
		cacheFile: path.Join(cwd, "testdata", "k3s-releases"),
		want: K3sVersionInfo{
			CacheFile:  path.Join(cwd, "testdata", "k3s-releases"),
			Versions:   []string{},
			ExpiryTime: time.Hour * 24,
			FromCache:  true,
		},
	})
	k3sVersionsTestCases.Set("expiredCache", k3sVersionsTestCase{
		cacheFile:  path.Join(cwd, "testdata", "k3s-releases-expired"),
		expiryTime: "1ms",
		want: K3sVersionInfo{
			CacheFile:  path.Join(cwd, "testdata", "k3s-releases-expired"),
			Versions:   []string{},
			ExpiryTime: time.Millisecond * 1,
			FromCache:  false,
		},
	})

	for _, key := range k3sVersionsTestCases.Keys() {
		name := key.(string)
		val, _ := k3sVersionsTestCases.Get(key)
		tc := val.(k3sVersionsTestCase)
		t.Run(name, func(t *testing.T) {
			if k3svi, err := NewK3sVersionUtil(tc.cacheFile, tc.expiryTime); err != nil {
				t.Errorf("Error %v", err)
			} else {
				if err := k3svi.QueryAndCacheK3sReleases(); err != nil {
					t.Errorf("Error %v", err)
				} else {
					if tc.want.FromCache != k3svi.FromCache {
						t.Errorf("Expecting FromCache to be %v but it is %v", tc.want.FromCache, k3svi.FromCache)
					}
					if tc.want.CacheFile != k3svi.CacheFile {
						t.Errorf("Expecting Cache File to be %s but it is %s", tc.want.CacheFile, k3svi.CacheFile)
					}
					if len(k3svi.Versions) <= 0 {
						t.Errorf("Expecting Versions to be greater then 0")
					}
					if tc.want.ExpiryTime != k3svi.ExpiryTime {
						t.Errorf("Expecting ExpiryTime to be %s but it is %s", tc.want.ExpiryTime, k3svi.ExpiryTime)
					}
				}
			}
		})
	}
	//Cleanup the test data files
	os.Remove(path.Join(cwd, "testdata", "k3s-releases"))
	os.Remove(path.Join(cwd, "testdata", "k3s-releases-expired"))
}
