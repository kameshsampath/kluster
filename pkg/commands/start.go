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
	"github.com/kameshsampath/kluster/pkg/model"
	"github.com/kameshsampath/kluster/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

type StartOptions struct {
	profile        string
	memory         string
	cpus           int
	cloudInitFile  string
	diskSize       string
	withKubeConfig bool
	k3sVersion     string
	k3sServerFlags []string
}

// BundleOptions implements Interface
var _ Command = (*StartOptions)(nil)

func (opts *StartOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.profile, "profile", "p", "cluster1", "The profile of the  multipass vm. The same value will be set as k3s Kube context profile.")
	cmd.Flags().StringVarP(&opts.memory, "memory", "m", "4g", "The memory to allocate to Virtual Machine.")
	cmd.Flags().IntVarP(&opts.cpus, "cpus", "c", 2, "Number of CPUs to allocate to Virtual Machine.")
	cmd.Flags().StringVarP(&opts.diskSize, "disk-size", "d", "40g", "The Virtual Machine Disk Size.")
	cmd.Flags().BoolVarP(&opts.withKubeConfig, "with-kube-config", "w", true, "Download and write the content of Kubeconfig to $KUBECONFIG.")
	cmd.Flags().StringVarP(&opts.k3sVersion, "k3s-version", "k", "", "The k3s version to use default is latest")
	cmd.Flags().StringArrayVarP(&opts.k3sServerFlags, "k3s-server-flags", "s", nil, "The extra k3s server options, these values will be set to INSTALL_EXEC variable")
}

func (opts *StartOptions) Validate(_ *cobra.Command, _ []string) error {
	// no validation
	return nil
}

func (opts *StartOptions) Execute(cmd *cobra.Command, _ []string) error {
	klusterList, err := utils.Klusters()
	if err != nil {
		log.Fatalf("Error getting existing Klusters %v", err)
		return err
	}
	var k *model.Kluster
	if k = klusterList.HasKluster(opts.profile); k == nil {
		out, err := opts.startKluster()
		if err != nil {
			log.Fatalf("Error creating kluster %s, %s", opts.profile, err)
			fmt.Fprintf(cmd.ErrOrStderr(), "Error creating kluster %s, %s", opts.profile, err)
			return err
		}
		log.Infof("Kluster %s sucessfully created %s", opts.profile, utils.ToString(out))
		defer os.RemoveAll(opts.cloudInitFile)
	}
	log.Infof("Kluster %s already exsits", opts.profile)
	if opts.withKubeConfig {
		if kcutil, err := utils.NewKubeConfigUtil(""); err != nil {
			log.Warnf("Error building KubeConfigUtil %v while trying to write kubeconfig for kluster %s", err, opts.profile)
		} else {
			kd, err := utils.KlusterDetails(k.Name)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error getting details for kluster %s", opts.profile)
			}

			k.IPAddresses = kd.Info.IPAddresses
			if err = kcutil.WriteKubeConfig(opts.profile, k); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing kubeconfig for kluster %s", opts.profile)
			}
		}
	}
	return nil
}

func (opts *StartOptions) startKluster() ([]string, error) {
	c, err := NewCloudInitUtil(nil)
	if err != nil {
		log.Errorf("Erorr loading Cloud Init file %v", err)
		return nil, err
	}
	if err := c.configureAndGetCloudInitFile(opts); err != nil {
		return nil, err
	}
	launchArgs := []string{"launch",
		fmt.Sprintf("--name=%s", opts.profile),
		fmt.Sprintf("--mem=%s", opts.memory),
		fmt.Sprintf("--cpus=%d", opts.cpus),
		fmt.Sprintf("--disk=%s", opts.diskSize),
		fmt.Sprintf("--cloud-init=%s", opts.cloudInitFile),
	}
	log.Debugf("Launching kluster with arguments %s", launchArgs)
	er := utils.NewExecUtil()
	if err := er.Execute(launchArgs); err != nil {
		instanceExistErr := fmt.Sprintf("launch failed: instance \"%s\" already exists", opts.profile)
		if len(er.StdOutErrLines) > 0 && instanceExistErr != er.StdOutErrLines[0] {
			log.Fatalf("Error creating cluster %s", er.StdErr.String())
			return nil, err
		}
	} else {
		defer os.RemoveAll(opts.cloudInitFile)
		return er.StdOutErrLines, nil
	}
	return nil, nil
}

var startCommandExample = fmt.Sprintf(`
  # Create kluster with default options
  %[1]s start --profile foo
  # Create kluster with custom memory and cpus
  %[1]s start --profile foo --memory=8g --cpus=4g
`, ExamplePrefix())

//NewStartCommand instantiates the new instance of the StartCommand
func NewStartCommand() *cobra.Command {
	startOpts := &StartOptions{}

	startCmd := &cobra.Command{
		Use:     "start",
		Short:   "Start a k3s cluster with multipass",
		Example: startCommandExample,
		RunE:    startOpts.Execute,
		PreRunE: startOpts.Validate,
	}

	startOpts.AddFlags(startCmd)

	return startCmd
}
