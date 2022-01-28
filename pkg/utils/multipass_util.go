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

package utils

import (
	"github.com/kameshsampath/kluster/pkg/model"
	log "github.com/sirupsen/logrus"
)

//Klusters lists all the available klusters
func Klusters() (*model.KlusterList, error) {
	commandArgs := []string{"list", "--format=json"}
	er := NewExecUtil()
	if err := er.Execute(commandArgs); err != nil {
		log.Errorf("Error getting kluster list %s, %s", er.StdErr.String(), err)
		return nil, err
	}
	var klusterList model.KlusterList
	if err := klusterList.MachineListToKlusterList(er.StdOut.Bytes()); err != nil {
		return nil, err
	}
	return &klusterList, nil
}

//KlusterDetails gets the details of the kluster named klusterName
func KlusterDetails(klusterName string) (*model.KlusterDetails, error) {
	commandArgs := []string{"info", klusterName, "--format=json"}
	er := NewExecUtil()
	if err := er.Execute(commandArgs); err != nil {
		log.Errorf("Error getting kluster %s details %s", klusterName, er.StdErr.String())
		return nil, err
	}
	var kd model.KlusterDetails
	if err := kd.MachineInfoToKlusterDetails(er.StdOut.Bytes()); err != nil {
		return nil, err
	}
	return &kd, nil
}
