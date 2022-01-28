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
	"github.com/kameshsampath/kluster/pkg/model"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	log "github.com/sirupsen/logrus"
	"gopkg.in/op/go-logging.v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"os"
	"path"
	"sigs.k8s.io/yaml"
	"strings"
)

func init() {
	// setting the yqlib logger to log from warning and above
	stdOutBackend := logging.NewLogBackend(os.Stdout, "", 0)
	stdOutBackendLeveled := logging.AddModuleLevel(stdOutBackend)
	stdOutBackendLeveled.SetLevel(logging.WARNING, "")
	yqlib.GetLogger().SetBackend(stdOutBackendLeveled)
}

type KubeConfigFile struct {
	FilePath string
	Config   *clientcmdapi.Config
}

//NewKubeConfigUtil instantiates a new utility to manipulate Kube Config
func NewKubeConfigUtil(filePath string) (*KubeConfigFile, error) {
	if filePath != "" {
		return &KubeConfigFile{
			FilePath: filePath,
		}, nil
	}
	if os.Getenv("KUBECONFIG") != "" {
		filePath = os.Getenv("KUBECONFIG")
	} else if os.Getenv("KUBECONFIG") == "" {
		filePath = path.Join(defaultKubeConfigFileName)
	}
	kubeConfigDir := path.Dir(filePath)
	if _, err := os.Stat(kubeConfigDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(kubeConfigDir, 0700); err != nil {
				return nil, err
			}
		} else {
			log.Errorf("Error accessing %s kubeconfig file", filePath)
			return nil, err
		}
	}
	return &KubeConfigFile{
		FilePath: filePath,
	}, nil
}

//WriteKubeConfig writes the kubeconfig content of the cluster klusterName to file kubeConfigFile
func (k *KubeConfigFile) WriteKubeConfig(profileName string, kluster *model.Kluster) error {
	commandArgs := []string{"exec",
		profileName,
		"cat",
		"/etc/rancher/k3s/k3s.yaml",
	}
	log.Debugf("Getting kubeconfig from kluster %s", profileName)
	er := NewExecUtil()
	if err := er.Execute(commandArgs); err != nil {
		log.Errorf("Error getting kubeconfig from kluster %s:\n%s\n%v", profileName, er.StdErr.String(), err)
		return err
	}
	log.Debugf("Retrieved kubeconfig from Kluster %s: %s", profileName, er.StdOut.String())
	if err := k.mergeConfigs(kluster, er.StdOut.Bytes()); err != nil {
		log.Errorf("Error writing kubeconfig file %v", err)
		return err
	}
	log.Infof("Wrote kubeconfig for Kluster %s sucessfully to file %s", profileName, k.FilePath)
	return nil
}

//mergeConfigs merges the kubeconfig content to the kubeConfigFile
func (k *KubeConfigFile) mergeConfigs(kluster *model.Kluster, b []byte) error {
	dir, err := ioutil.TempDir(os.TempDir(), "kluster-kubeconfigs")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	_, err = os.ReadFile(k.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.OpenFile(k.FilePath, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				return err
			}
			log.Debugf("Kube Config file %v is empty, writing current content to it", k)
			_, err = f.Write(b)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	newConfig, err := ioutil.TempFile(dir, "new.config")
	if err != nil {
		return err
	}
	defer os.Remove(newConfig.Name())
	if _, err := newConfig.Write(b); err != nil {
		return err
	}
	if err := k.LoadKubeConfig(k.FilePath, newConfig.Name()); err != nil {
		return err
	}
	k.updateKubeClusterNameAndIP(kluster)
	if err = k.WriteToFile(); err != nil {
		return err
	}
	return nil
}

//LoadKubeConfig load the Kubeconfig file
func (k *KubeConfigFile) LoadKubeConfig(configFiles ...string) error {
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		WarnIfAllMissing: true,
		//TODO add -kubeconfig flag ?
		Precedence: configFiles,
	}
	load, err := loadingRules.Load()
	if err != nil {
		return err
	}
	k.Config = load

	return nil
}

func (k *KubeConfigFile) WriteToFile() error {
	json, err := runtime.Encode(clientcmdlatest.Codec, k.Config)
	if err != nil {
		log.Errorf("Unexpected error while writing Kubeconfig %s: %v", k.FilePath, err)
		return err
	}

	output, err := yaml.JSONToYAML(json)
	if err != nil {
		log.Errorf("Unexpected error while writing Kubeconfig %s: %v", k.FilePath, err)
		return err
	}

	if err := ioutil.WriteFile(k.FilePath, output, 0600); err != nil {
		log.Errorf("Error writing kubeconfig %s,%v", k.FilePath, err)
		return err
	}
	return nil
}

func (k *KubeConfigFile) updateKubeClusterNameAndIP(kluster *model.Kluster) {
	config := k.Config
	if v, ok := config.Clusters["default"]; ok {
		v.Server = strings.ReplaceAll(config.Clusters["default"].Server, "127.0.0.1", kluster.IPAddresses[0])
		config.Clusters[kluster.Name] = v
		delete(config.Clusters, "default")
	}
	if v, ok := config.AuthInfos["default"]; ok {
		v.Username = kluster.Name
		config.AuthInfos[kluster.Name] = v
		delete(config.AuthInfos, "default")
	}
	if v, ok := config.Contexts["default"]; ok {
		v.Cluster = kluster.Name
		v.AuthInfo = kluster.Name
		config.Contexts[kluster.Name] = v
		delete(config.Contexts, "default")
	}
	config.CurrentContext = kluster.Name
}

//RemoveEntriesFromKubeConfig Removes the entries from kubeconfig clusters etc.,
func (k *KubeConfigFile) RemoveEntriesFromKubeConfig(klusterName string) error {
	if err := k.LoadKubeConfig(k.FilePath); err != nil {
		log.Errorf("Error loading kubeconfig from file %s,%v", k.FilePath, err)
		return err
	}
	delete(k.Config.Clusters, klusterName)
	delete(k.Config.Contexts, klusterName)
	delete(k.Config.AuthInfos, klusterName)

	if err := k.WriteToFile(); err != nil {
		log.Errorf("Error updating to kubeconfig file %s,%v", k.FilePath, err)
		return err
	}

	return nil
}
