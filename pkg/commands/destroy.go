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
	utils "github.com/kameshsampath/go-kluster/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

type DestroyOptions struct {
	profile           string
	removeKubeContext bool
}

// BundleOptions implements Interface
var _ Command = (*DestroyOptions)(nil)

func (opts *DestroyOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.profile, "profile", "p", "", "The profile of the  multipass vm. The same value will be set as k3s Kube context profile.")
	if err := cmd.MarkFlagRequired("profile"); err != nil {
		log.Fatalf("Error marking flag 'profile' as requried %v", err)
	}
	cmd.Flags().BoolVarP(&opts.removeKubeContext, "remove-from-kube-context", "r", false, "Remove the Kluster kube context entry from $KUBECONFIG.")
}

func (opts *DestroyOptions) Validate(cmd *cobra.Command, args []string) error {
	viper.BindPFlags(cmd.Flags())

	if opts.profile = viper.GetString("profile"); opts.profile == "" {
		return fmt.Errorf("kluster's 'profile' is required")
	}

	return nil
}

func (opts *DestroyOptions) Execute(cmd *cobra.Command, args []string) error {
	log.Debugf("Deleting kluster %s", opts.profile)
	if out, err := opts.destroyKluster(); err != nil {
		kerr := fmt.Sprintf("kluster \"%s\" does not exist", opts.profile)
		if err.Error() == kerr {
			fmt.Println(cmd.OutOrStdout(), kerr)
		} else {
			log.Fatalf(`Error deleting kluster %s
%s,%v`, opts.profile, utils.ToString(out), err)
		}
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Kluster %s sucessfully deleted.", opts.profile)
	}
	opts.deleteKubeContextFromKubeConfig()
	return nil
}

func (opts *DestroyOptions) destroyKluster() ([]string, error) {
	klusterList, err := utils.Klusters()
	if err != nil {
		log.Errorf("Error getting Klusters %v", err)
		return nil, err
	}
	if klusterList.HasKluster(opts.profile) != nil {
		deleteArgs := []string{"delete",
			opts.profile,
			"--purge",
		}
		log.Debugf("Command arguments %s", strings.Join(deleteArgs, " "))
		er := utils.NewExecUtil()
		if err := er.Execute(deleteArgs); err != nil {
			return nil, err
		}
		log.Debugf("Kluster %s sucessfully deleted %s", opts.profile, er.StdErr.String())
		opts.deleteKubeContextFromKubeConfig()
		return er.StdOutErrLines, nil
	}
	return nil, fmt.Errorf("kluster \"%s\" does not exist", opts.profile)
}

func (opts *DestroyOptions) deleteKubeContextFromKubeConfig() {
	if opts.removeKubeContext {
		log.Infof("Deleting Kluster %s kubecontext entry", opts.profile)
		if kcutil, err := utils.NewKubeConfigUtil(""); err != nil {
			log.Warnf("Error building KubeConfigUtil %v while trying to write kubeconfig for kluster %s", err, opts.profile)
		} else {
			if err := kcutil.RemoveEntriesFromKubeConfig(opts.profile); err != nil {
				log.Warnf("Unable to remove %s entry from kubeconfig file %#v", opts.profile, err)
			}
		}
	}
}

var destroyCommandExample = fmt.Sprintf(`
  # Delete kluster with default options
  %[1]s destroy --profile foo
  # Create kluster and remove the kluster entry from kubeconfig
  %[1]s destroy --profile foo --remove-from-kube-context
`, ExamplePrefix())

//NewDestroyCommand instantiates the new instance of the destroy commands
func NewDestroyCommand() *cobra.Command {
	destroyOpts := &DestroyOptions{}

	destroyCommand := &cobra.Command{
		Use:     "destroy",
		Short:   "Destroy an existing kluster.",
		Example: destroyCommandExample,
		RunE:    destroyOpts.Execute,
		PreRunE: destroyOpts.Validate,
	}

	destroyOpts.AddFlags(destroyCommand)

	return destroyCommand
}
