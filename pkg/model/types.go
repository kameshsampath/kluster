/*
Copyright © 2022 Kamesh Sampath <kamesh.sampath@hotmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
)

// Memory holds the memory allocated to the Kluster
type Memory struct {
	Total float64 `json:"total"`
	Used  float64 `json:"used"`
}

//Kluster holds the information about the Kluster
type Kluster struct {
	IPAddresses []string `json:"ipv4"`
	Name        string   `json:"name"`
	Release     string   `json:"release"`
	State       string   `json:"state"`
	Memory      Memory   `json:"memory,omitempty"`
}

//KlusterList holds the list of Klusters i.e. multipass vm as list
type KlusterList struct {
	Klusters []Kluster `json:"list"`
}

//KlusterInfo is type of Kluster to be used while processing the machine details
type KlusterInfo Kluster

//KlusterDetails holds the details of the kluster
type KlusterDetails struct {
	Info   KlusterInfo `json:"info"`
	Errors []string    `json:"errors"`
}

var _ sort.Interface = (*KlusterList)(nil)

//UnmarshalJSON helps in custom marshalling of KlusterInfo as the
// json from the commands output has dynamic field with name of the kluster
func (k *KlusterInfo) UnmarshalJSON(data []byte) error {
	var vmInfo interface{}
	if err := json.Unmarshal(data, &vmInfo); err != nil {
		return err
	}
	m := vmInfo.(map[string]interface{})
	for n, v := range m {
		k.Name = n
		t := v.(map[string]interface{})
		if rel, ok := t["image_release"]; ok {
			k.Release = rel.(string)
		} else if rel, ok := t["release"]; ok {
			k.Release = rel.(string)
		}
		k.State = t["state"].(string)
		if t["ipv4"] != nil {
			for _, y := range t["ipv4"].([]interface{}) {
				k.IPAddresses = append(k.IPAddresses, y.(string))
			}
		}
		if mem, ok := t["memory"]; ok {
			memMap := mem.(map[string]interface{})
			var mem Memory
			if tm, ok := memMap["total"]; ok {
				mem.Total = tm.(float64)
			}
			if um, ok := memMap["used"]; ok {
				mem.Used = um.(float64)
			}
			k.Memory = mem
		}
		// since this map consist of only one element we can return immediately
		return nil
	}
	return nil
}

func (k KlusterInfo) String() string {
	return fmt.Sprintf("Name %s, Release: %s, State: %s", k.Name, k.Release, k.State)
}

func (k Kluster) String() string {
	return fmt.Sprintf("Name %s, Release: %s, State: %s", k.Name, k.Release, k.State)
}

func (k Kluster) IP() string {
	if len(k.IPAddresses) > 0 {
		return k.IPAddresses[0]
	}
	return ""
}

//HasKluster checks if the list has the kluster by name klusterName
func (kl *KlusterList) HasKluster(klusterName string) *Kluster {
	for _, e := range kl.Klusters {
		if e.Name == klusterName {
			return &e
		}
	}
	return nil
}

//GetIP get the ip address of the Kluster vm
func (k Kluster) GetIP(index int) string {
	if len(k.IPAddresses) > 0 {
		return k.IPAddresses[index]
	}
	return ""
}

func (kl *KlusterList) MachineListToKlusterList(out []byte) error {
	log.Debugf("Json Response %s", string(out))
	if err := json.Unmarshal(out, &kl); err != nil {
		log.Errorf("Error loading list of klusters %s", err)
		return nil
	}
	return nil
}

func (klusterDetails *KlusterDetails) MachineInfoToKlusterDetails(out []byte) error {
	log.Debugf("Json Response %s", string(out))
	if err := json.Unmarshal(out, &klusterDetails); err != nil {
		log.Errorf("Error loading list of klusters %s", err)
		return nil
	}
	return nil
}

func (kl *KlusterList) Len() int {
	return len(kl.Klusters)
}

func (kl KlusterList) Less(i, j int) bool {
	return kl.Klusters[i].Name < kl.Klusters[j].Name
}

func (kl KlusterList) Swap(i, j int) {
	kl.Klusters[i], kl.Klusters[j] = kl.Klusters[j], kl.Klusters[i]
}
