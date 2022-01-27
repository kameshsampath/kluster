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
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

var baseCommand = "multipass"

type ExecResponse struct {
	Command        string
	StdOut         bytes.Buffer
	StdErr         bytes.Buffer
	StdOutErrLines []string
}

//NewExecUtil returns a new exec utility to exec native process
func NewExecUtil() *ExecResponse {
	return &ExecResponse{}
}

//Execute executes the commands with launchArgs as parameters
func (execResponse *ExecResponse) Execute(commandArgs []string) error {
	//some cases you may need to do sudo for the command
	//in those cases use the KLUSTER_WITH_SUDO to run the multipass with sudo
	if os.Getenv("KLUSTER_WITH_SUDO") != "" {
		log.Debugf("Ensuring the command is run with sudo")
		commandArgs = append([]string{"-E", baseCommand}, commandArgs...)
		baseCommand = "sudo"
	}
	cmd := exec.Command(baseCommand, commandArgs...)
	execResponse.Command = cmd.String()
	log.Debugf("Executing command %s", execResponse.Command)
	cmd.Stdout = &execResponse.StdOut
	cmd.Stderr = &execResponse.StdErr
	if err := cmd.Run(); err != nil {
		log.Errorf("Error running command %s, %v", execResponse.Command, err)
		execResponse.StdOutErrLines = sanitize(execResponse.StdErr)
		return err
	}
	log.Debugf("Command %s successfully executed", execResponse.Command)
	execResponse.StdOutErrLines = sanitize(execResponse.StdOut)
	return nil
}

//sanitize cleans up the line removing extra newlines and control characters
func sanitize(buf bytes.Buffer) []string {
	if str, _ := buf.ReadString('\n'); str != "" {
		lines := strings.Split(str, "\n")
		for i, l := range lines {
			if l != "" {
				s := strings.TrimSpace(l)
				s = strings.ReplaceAll(s, "\n", "")
				lines[i] = s
			}
		}
		return lines
	}
	return nil
}
