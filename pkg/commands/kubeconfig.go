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

package commands

import (
	"fmt"
	utils "github.com/kameshsampath/kluster/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type KubeConfigOptions struct {
	profile        string
	kubeConfigFile string
}

// KubeConfigOptions implements Interface
var _ Command = (*KubeConfigOptions)(nil)

func (opts *KubeConfigOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.profile, "profile", "p", "", "The profile of the Kluster from which extract the kube config")
	if err := cmd.MarkFlagRequired("profile"); err != nil {
		log.Fatalf("Error marking flag 'profile' as requried %v", err)
	}
	cmd.Flags().StringVarP(&opts.kubeConfigFile, "to-file", "f", "", "The file to which the contents of the kubeconfig will be written")
}

func (opts *KubeConfigOptions) Validate(cmd *cobra.Command, args []string) error {
	// no validation
	return nil
}

func (opts *KubeConfigOptions) Execute(cmd *cobra.Command, args []string) error {
	klusterList, err := utils.Klusters()
	if err != nil {
		log.Fatalf("Error getting Klusters %v", err)
		return err
	}
	if klusterList.HasKluster(opts.profile) != nil {
		kcutil, err := utils.NewKubeConfigUtil(opts.kubeConfigFile)
		if err != nil {
			log.Fatalf("Error building KubeConfigUtil %v while trying to write kubeconfig for kluster %s", err, opts.profile)
			return err
		}
		if err := kcutil.WriteKubeConfig(opts.profile, nil); err != nil {
			return err
		}
	}
	log.Infof("Kluster %s does not exist", opts.profile)
	return nil
}

var kubeConfigCommandExample = fmt.Sprintf(`
  # Get Kube Config for kluster and write contents to $KUBECONFIG
  %[1]s kubeconfig --profile foo
  # Get Kube Config for kluster and write contents to file named /home/foo/my-kube-config-file
  %[1]s kubeconfig --profile foo --to-file /home/foo/my-kube-config-file
`, ExamplePrefix())

//NewKubeConfigCommand instantiates the new instance of the kubeconfig
func NewKubeConfigCommand() *cobra.Command {
	kcOpts := &KubeConfigOptions{}

	kubeConfigCommand := &cobra.Command{
		Use:     "kubeconfig",
		Short:   "Get the Kube config from the kluster and write the contents to file $KUBECONFIG",
		Example: kubeConfigCommandExample,
		RunE:    kcOpts.Execute,
		PreRunE: kcOpts.Validate,
	}

	kcOpts.AddFlags(kubeConfigCommand)

	return kubeConfigCommand
}
