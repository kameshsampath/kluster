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
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

//ReadFile reads the contents of the file src as list of lines
func ReadFile(src string) ([]string, error) {
	file, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	defer file.Close()
	return lines, nil
}

//WriteLinesToFile write the list of lines to file dest
func WriteLinesToFile(dest string, lines []string) error {
	file, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Errorf("Error writing content to file %s, %v", dest, err)
		return err
	}
	fileWriter := bufio.NewWriter(file)
	for _, line := range lines {
		_, _ = fileWriter.WriteString(fmt.Sprintf("%s\n", line))
	}
	fileWriter.Flush()
	defer file.Close()
	return nil
}

//ToString joins the lines seperated by newline
func ToString(lines []string) string {
	var str string
	for _, e := range lines {
		str = str + fmt.Sprintf("%s\n", e)
	}
	return str
}
