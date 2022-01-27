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
	"github.com/kameshsampath/go-kluster/pkg/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type User struct {
	Name              string   `yaml:"profile,omitempty"`
	Groups            string   `yaml:"groups,omitempty"`
	Shell             string   `yaml:"shell,omitempty"`
	Sudo              string   `yaml:"sudo,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh-authorized-keys,omitempty"`
	Password          string   `yaml:"password,omitempty"`
}

type CloudInit struct {
	PackageUpdate bool                  `yaml:"package_update,omitempty"`
	Packages      []string              `yaml:"packages,omitempty"`
	Bootcmd       []string              `yaml:"bootcmd,omitempty"`
	Runcmd        []string              `yaml:"runcmd,omitempty"`
	Users         []User                `yaml:"users"`
	K3sUtil       *utils.K3sVersionInfo `yaml:"-"`
}

var _ yaml.Unmarshaler = (*User)(nil)

func NewCloudInitUtil(k3sUtil *utils.K3sVersionInfo) (*CloudInit, error) {
	var err error
	if k3sUtil == nil {
		k3sUtil, err = utils.NewK3sVersionUtil("", "")
		if err != nil {
			return nil, err
		}
	}
	if err := k3sUtil.QueryAndCacheK3sReleases(); err != nil {
		return nil, err
	}
	return &CloudInit{
		K3sUtil: k3sUtil,
	}, nil
}

//UnmarshalYAML implements Unmarshaller to unmarshall User YAML data
func (u *User) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t interface{}
	if err := unmarshal(&t); err != nil {
		return err
	}
	if s, ok := t.(string); ok {
		u.Name = s
	} else if m, ok := t.(map[interface{}]interface{}); ok {
		for k, v := range m {
			switch key := k.(string); key {
			case "profile":
				u.Name = v.(string)
			case "groups":
				u.Groups = v.(string)
			case "password":
				u.Password = v.(string)
			case "shell":
				u.Shell = v.(string)
			case "sudo":
				u.Sudo = v.(string)
			case "ssh-authorized-keys":
				iArr := v.([]interface{})
				var sshAuthKeys []string
				for _, i := range iArr {
					sshAuthKeys = append(sshAuthKeys, i.(string))
				}
				u.SSHAuthorizedKeys = sshAuthKeys
			}
		}
	}
	return nil
}

//buildCloudInitConfig marshalls cloud init YAML to CloudInit
func (c *CloudInit) configureAndGetCloudInitFile(opts *StartOptions) error {
	log.Debugf("Runner Temp is %s and user temp dir is %s ", os.Getenv("TMPDIR"), os.TempDir())
	var b = []byte(k3sDefaultCloudInit)
	if err := yaml.Unmarshal(b, c); err != nil {
		log.Errorf("Error unmarshalling Cloud Init %v", err)
		return err
	}
	if opts.k3sVersion == "" {
		opts.k3sVersion = c.K3sUtil.Versions[0]
	}
	installExec := fmt.Sprintf(k3sInstallCmd, opts.k3sVersion, strings.Join(opts.k3sServerFlags, " "))
	c.Runcmd = append(c.Runcmd, installExec, k3sKubeConfigCopyCmd)
	dir, err := ioutil.TempDir("", "kluster-start")
	if err != nil {
		return err
	}
	cloudInitFile := path.Join(dir, fmt.Sprintf("%s-cloud-init", opts.profile))
	if _, err := os.Create(cloudInitFile); err != nil {
		log.Errorf("Unable to create  cloud init file %s, %v", cloudInitFile, err)
		return err
	}
	b, err = yaml.Marshal(&c)
	if err != nil {
		log.Errorf("Error marshalling cloud init file %v", err)
		return err
	}
	if err := ioutil.WriteFile(cloudInitFile, b, 0600); err != nil {
		log.Errorf("Error writing to cloud init file %v", err)
		return err
	}
	log.Debugf("Generated clou-init file %s", cloudInitFile)
	opts.cloudInitFile = cloudInitFile
	return nil
}
