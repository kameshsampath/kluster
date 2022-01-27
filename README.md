# Kluster

Kluster is a tool used to build and run [k3s](https://k3s.io) clusters. The tool leverages [multipass](https://multipass.run) for creating a Virtual Machine to host and run the k3s cluster.

__NOTE__: The tool is under early stages of development and lots of changes underway. 

## Install

Download and install [multipass](https://multipass.run) and then grab the binary from [releases](https://github.com/kameshsampath/kluster/releases).

## Usage

### Create

```shell
kluster start --profile=foo
```

### Get Kubeconfig

```shell
kluster kubeconfig --profile=foo
```

### Delete

```shell
kluster delete --profile=foo
```

### Help

```shell
kluster --help
```
