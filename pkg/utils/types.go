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

import "fmt"

const (
	yqClustersCountExpr       = ".clusters|length"
	yqClustersNameExpr        = ".clusters.[].name"
	yqContextsCountExpr       = ".contexts|length"
	yqContextsNameExpr        = ".contexts.[].name"
	yqUsersCountExpr          = ".users|length"
	yqUsersNameExpr           = ".users.[].name"
	yqExtensionsCountExpr     = ".extensions|length"
	yqPrefExtensionsCountExpr = ".preferences.extensions|length"
	mergeExpr                 = "select(fileIndex == 0) *+ select(fileIndex == 1)"
	defaultKubeConfigFileName = "config"
)

var kubeIPNameReplaceExpr = `(.clusters.[].cluster.server | select(. == "https://127.0.0.1:6443")) = "https://%s:6443" | 
(.clusters.[].name | select(. == "default")) = "%[2]s" |
(.contexts.[].name | select(. == "default")) = "%[2]s" | 
(.contexts.[].context.cluster | select(. == "default")) = "%[2]s" | 
(.contexts.[].context.user | select(. == "default")) = "%[2]s" | 
(.users.[].name | select(. == "default")) = "%[2]s"`

//GHRelease struct to hold the GH release info like tag and release type
type GHRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft,omitempty"`
	PreRelease bool   `json:"prerelease,omitempty"`
}

func deleteKubeContextExpressionForKluster(klusterName string) string {
	return fmt.Sprintf(`del(.clusters.[] | select(.name == "%[1]s")) | 
del(.users.[] | select(.name == "%[1]s"))  | del(.contexts.[] |
select(.name == "%[1]s"))| del(.extensions.[] | select(.name == "%[1]s")) | 
del(.preferences.extensions.[] | select(.name == "%[1]s")) | .current-context |= ""
`, klusterName)
}
