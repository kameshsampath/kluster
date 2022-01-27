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
	"bytes"
	"fmt"
	"github.com/kameshsampath/kluster/pkg/model"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	log "github.com/sirupsen/logrus"
	"gopkg.in/op/go-logging.v1"
	"io/ioutil"
	"os"
	"path"
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
		filePath = path.Join(os.Getenv("HOME"), ".kube", defaultKubeConfigFileName)
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
	if err := k.mergeConfigs(kluster, er.StdOutErrLines); err != nil {
		log.Errorf("Error writing kubeconfig file %v", err)
		return err
	}
	log.Infof("Wrote kubeconfig for Kluster %s sucessfully to file %s", profileName, k.FilePath)
	return nil
}

//mergeConfigs merges the kubeconfig content to the kubeConfigFile
func (k *KubeConfigFile) mergeConfigs(kluster *model.Kluster, kubeConfigLines []string) error {
	dir, err := ioutil.TempDir(os.TempDir(), "kluster-kubeconfigs")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	lines, err := ReadFile(k.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.Create(k.FilePath)
			log.Debugf("Kube Config file %s is empty, writing current content to it", k)
			err = WriteLinesToFile(k.FilePath, kubeConfigLines)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if lines == nil {
		log.Debugf("Kube Config file %s is empty, writing current content to it", k)
		err = WriteLinesToFile(k.FilePath, kubeConfigLines)
		if err != nil {
			return err
		}
	}
	newConfig, err := ioutil.TempFile(dir, "new.config")
	if err != nil {
		return err
	}
	defer os.RemoveAll(newConfig.Name())
	if err := WriteLinesToFile(newConfig.Name(), kubeConfigLines); err != nil {
		return err
	}

	outBuffer := new(bytes.Buffer)
	hasCluster := false
	kubeConfigFiles := []string{k.FilePath, newConfig.Name()}

	clusterNames, err := Evaluate(k.FilePath, ".clusters.[].name")
	if err != nil {
		return err
	}
	for _, n := range clusterNames {
		if n == kluster.Name {
			hasCluster = true
			break
		}
	}
	//Update the kubeconfig only if the kubeconfig does not have the context
	//TODO replace ??
	if !hasCluster {
		pw := yqlib.NewPrinterWithSingleWriter(outBuffer, yqlib.YamlOutputFormat, true, false, 2, false)
		if err := yqlib.NewAllAtOnceEvaluator().EvaluateFiles(
			mergeExpr, kubeConfigFiles,
			pw, false); err != nil {
			return err
		}
		//Write the merged content to the k
		if err := ioutil.WriteFile(k.FilePath, outBuffer.Bytes(), 0600); err != nil {
			return err
		}

		if err := k.updateKubeClusterNameAndIP(kluster); err != nil {
			log.Errorf("Error updating kubeconfig file with IP and Name, %v", err)
			return err
		}
	}
	return nil
}

func (k *KubeConfigFile) updateKubeClusterNameAndIP(kluster *model.Kluster) error {
	expr := fmt.Sprintf(kubeIPNameReplaceExpr, kluster.IPAddresses[0], kluster.Name)
	buf := new(bytes.Buffer)
	printer := yqlib.NewPrinterWithSingleWriter(buf, yqlib.YamlOutputFormat,
		true, false, 2, false)
	if err := yqlib.NewStreamEvaluator().EvaluateFiles(expr,
		[]string{k.FilePath}, printer, false); err != nil {
		return err
	}
	if err := ioutil.WriteFile(k.FilePath, buf.Bytes(), 0600); err != nil {
		return err
	}
	log.Infof("Successfully updated the Kube server IP and name for kluster %s", kluster.Name)
	return nil
}

//Evaluate will evaluate YAML expressions on kubeConfigFile and return a response of the evaluation
func Evaluate(kubeConfigFile, expr string) ([]string, error) {
	buf := new(bytes.Buffer)
	printer := yqlib.NewPrinterWithSingleWriter(buf, yqlib.YamlOutputFormat,
		true, false, 2, false)
	if err := yqlib.NewStreamEvaluator().EvaluateFiles(expr,
		[]string{kubeConfigFile}, printer, false); err != nil {
		return []string{}, err
	}
	return strings.Fields(buf.String()), nil
}

//RemoveEntriesFromKubeConfig Removes the entries from kubeconfig clusters etc.,
func (k *KubeConfigFile) RemoveEntriesFromKubeConfig(klusterName string) error {
	outBuffer := new(bytes.Buffer)
	pw := yqlib.NewPrinterWithSingleWriter(outBuffer, yqlib.YamlOutputFormat, true, false, 2, false)
	if err := yqlib.NewAllAtOnceEvaluator().EvaluateFiles(
		deleteKubeContextExpressionForKluster(klusterName), []string{k.FilePath},
		pw, false); err != nil {
		return err
	}
	//Write the update content to the k
	if err := ioutil.WriteFile(k.FilePath, outBuffer.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}
