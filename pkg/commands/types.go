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

package commands

const (
	k3sInstallCmd        = `curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="%s" INSTALL_K3S_EXEC="%s" K3S_KUBECONFIG_MODE="644" sh -s -`
	k3sKubeConfigCopyCmd = `mkdir -p /home/ubuntu/.kube && cp /etc/rancher/k3s/k3s.yaml /home/ubuntu/.kube/config`
	k3sDefaultCloudInit  = `
#cloud-config
package_update: true
packages:
  - net-tools
  - traceroute
  - arping
  - bridge-utils
  - jq
bootcmd:
  - sysctl -w net.ipv4.ip_forward=1
  - sysctl -w net.ipv6.conf.all.forwarding=1
  - sysctl -p
runcmd:
  - 'chown -R ubuntu:ubuntu /home/ubuntu/.kube'
  - 'echo "source <(kubectl completion bash)" >> /home/ubuntu/.bashrc'
  - 'echo "alias k=kubectl" >> /home/ubuntu/.bashrc'
  - 'echo "complete -F __start_kubectl k" >> /home/ubuntu/.bashrc'
  - 'echo "[ -f ~/.kubectl_aliases ] && source ~/.kubectl_aliases" >> /home/ubuntu/.bashrc'
  - 'curl -L https://raw.githubusercontent.com/ahmetb/kubectl-aliases/master/.kubectl_aliases -o /home/ubuntu/.kubectl_aliases'
  - 'curl -L https://raw.githubusercontent.com/ahmetb/kubectx/master/kubectx -o /usr/local/bin/kubectx'
  - 'curl -L https://raw.githubusercontent.com/ahmetb/kubectx/master/kubens -o /usr/local/bin/kubens'
  - 'chmod +x /usr/local/bin/kubectx /usr/local/bin/kubens'
  - 'chown -R ubuntu:ubuntu /home/ubuntu/.kubectl_aliases'
users:
  - default
  - name: ubuntu
    groups: sudo
    shell: /bin/bash
    sudo: ALL=(ALL) NOPASSWD:ALL
    ssh-authorized-keys:
    # vagrant insecure public key replace with our own
    - "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzIw+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoPkcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NOTd0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcWyLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQ== vagrant insecure public key"
    # password generated using python mkpasswd.py with password ubuntu
    password: "$6$rounds=656000$bxo/LO22lvUl4WCy$s20dy/HYcmb.EOvusWCHBA8zm03VZun/846nmzBQ9S7BxNt7N5v4XX/Dh5CHEuQ0phbRENqtMdN7AXZaraKaB/"
`
)
