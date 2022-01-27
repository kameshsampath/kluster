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
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"regexp"
	"sort"
	"time"
)

const regx = `^(v[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})([\+|\-]k3s[0-9])$`

type K3sVersionInfo struct {
	Versions   []string
	CacheFile  string
	FromCache  bool
	ExpiryTime time.Duration
}

//NewK3sVersionUtil returns a new instance of K3sVersionInfo
func NewK3sVersionUtil(cacheFile string, expiryTime string) (*K3sVersionInfo, error) {
	var k3sVersionInfo K3sVersionInfo
	if cacheFile != "" {
		k3sVersionInfo.CacheFile = cacheFile
	} else {
		k3sVersionInfo.CacheFile = path.Join(os.Getenv("HOME"), ".kluster", "cache", "k3s-releases")
	}
	if expiryTime != "" {
		d, err := time.ParseDuration(expiryTime)
		if err != nil {
			return nil, err
		}
		k3sVersionInfo.ExpiryTime = d
	} else {
		k3sVersionInfo.ExpiryTime = time.Hour * 24
	}
	return &k3sVersionInfo, nil
}

//QueryAndCacheK3sReleases queries the k3s releases via GitHub API and cache the release
// in a file $HOME/.kluster/cache/k3s-releases
func (k3sv *K3sVersionInfo) QueryAndCacheK3sReleases() error {
	fi, err := os.Stat(k3sv.CacheFile)
	if err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(k3sv.CacheFile), 0700); err != nil {
			log.Errorf("Error kluster home directory %s:  %v", path.Dir(k3sv.CacheFile), err)
		}
		log.Infof("k3s releases are not in cache Querying it via API")
		if err = k3sv.queryAndCacheReleases(); err != nil {
			return err
		}
		return err
	} else if time.Now().After(fi.ModTime().Add(k3sv.ExpiryTime)) { //Refresh the cache if it's greater than a day
		log.Info("Refreshing k3s releases Cache")
		//clean old cache content after expiry
		if _, err := os.Create(k3sv.CacheFile); err != nil {
			log.Errorf("Error creating Cachhe file %s, %v", k3sv.CacheFile, err)
		} else {
			if err = k3sv.queryAndCacheReleases(); err != nil {
				return err
			}
		}
		return err
	} else {
		log.Infof("Loading k3s releases from Cache")
		lines, err := ReadFile(k3sv.CacheFile)
		if err != nil {
			log.Errorf("Error reading Cache file %s creating afresh, %v", k3sv.CacheFile, err)
			err = k3sv.queryAndCacheReleases()
			if err != nil {
				return err
			}
		}
		k3sv.FromCache = true
		var releases []string
		releases = append(releases, lines...)
		k3sv.Versions = releases
		return nil
	}
}

func (k3sv *K3sVersionInfo) queryAndCacheReleases() error {
	client := resty.New()
	resp, err := client.R().
		EnableTrace().
		SetHeader("Accept", "application/vnd.github.v3+json").
		Get("https://api.github.com/repos/k3s-io/k3s/releases")

	if err != nil {
		return err
	}
	var ghReleases []GHRelease
	if resp.IsSuccess() {
		err = json.Unmarshal(resp.Body(), &ghReleases)
		if err != nil {
			log.Errorf("Error querying releass %v", err)
		}
		log.Debugf("Releases %v", ghReleases)
		var releases []string
		for _, r := range ghReleases {
			if r.isNotPreOrDraftOrCandidateRelease() {
				releases = append(releases, r.TagName)
			}
		}
		sortAndReverse(releases)
		err = WriteLinesToFile(k3sv.CacheFile, releases)
		if err != nil {
			defer os.Remove(k3sv.CacheFile)
		}
		k3sv.Versions = releases
		return nil
	}
	return errors.New(string(resp.Body()))
}

func sortAndReverse(releases []string) {
	sort.Strings(releases)
	for i, j := 0, len(releases)-1; i < j; i, j = i+1, j-1 {
		releases[i], releases[j] = releases[j], releases[i]
	}
}

func (r GHRelease) isNotPreOrDraftOrCandidateRelease() bool {
	if !(r.PreRelease && r.Draft) {
		re, err := regexp.Compile(regx)
		if err != nil {
			return false
		}
		return re.Match([]byte(r.TagName))
	}
	return false
}
