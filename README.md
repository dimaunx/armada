# Armada

Armada is a tool for creating multiple k8s clusters with [kind] (k8s in docker). This tool relies heavily on [kind] and 
extends its functionality with automation to create clusters tailored for multi cluster/multi cni local development and testing.

[![Build Status](https://travis-ci.com/dimaunx/armada.svg?branch=master)](https://travis-ci.com/dimaunx/armada)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimaunx/armada)](https://goreportcard.com/report/github.com/dimaunx/armada)
[![GitHub release](https://img.shields.io/github/release/dimaunx/aramada.svg)](https://GitHub.com/dimaunx/armada/releases/)

## Prerequisites

- [go 1.12] with [$GOPATH configured]
- [docker]

#### Get the latest version from [Releases] page.


## Build the tool locally.

```bash
make build
```

## Build in docker.

```bash
make docker-build
```

The **armada** binary will be placed under local **./bin** directory.

## Create clusters

In order to run more then 3 clusters, the following limits must be increased:

```bash
echo fs.file-max=500000 | sudo tee -a /etc/sysctl.conf                                                                      
echo fs.inotify.max_user_instances=8192 | sudo tee -a /etc/sysctl.conf
echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf
sudo sysctl -p 
```

The tool will create 2 clusters by default with kindnet cni plugin.

```bash
cd ./bin
./armada create clusters
``` 

This command will create five clusters with default kindnet cni.

```bash
./armada create clusters -n 5
```

Create a total of four clusters, 2 with weave, one with flannel and one with calico.

```bash
./armada create clusters -n 2 --weave
./armada create clusters -n 3 --flannel
./armada create clusters -n 4 --calico
```

Default kubernetes node image is kindest/node:v1.15.3. To use different image use **-i** or **--image** flag. This command will create three clusters with flannel cni and kubernetes 1.14.6.

```bash
./armada create clusters -n 3 --flannel --image kindest/node:v1.14.6
```

Full list of supported images can be found on [kind release page].

Example of running four clusters with multiple k8s versions and different cni plugins.

```bash
./armada create clusters -n 2 --weave  # 2 clusters with weave, k8s version 1.15.3
./armada create clusters -n 3 --flannel --image kindest/node:v1.14.6 # one clusters with flannel cni, k8s version 1.14.6
./armada create clusters -n 4 --calico --image kindest/node:v1.13.10 # one clusters with calico cni, k8s version 1.13.10
```

Alternatively run in docker after running **make docker-build** or **make build** commands. This command will create four cluster with calico cni.

```bash
make docker-run ARGS="./armada create clusters -n 4 --calico"
``` 

## Destroy clusters

```bash
./armada destroy clusters
``` 

Destroy clusters from docker.

```bash
make docker-run ARGS="./armada destroy clusters"
``` 

<!--links-->
[go 1.12]: https://blog.golang.org/go1.12
[docker]: https://docs.docker.com/install/
[$GOPATH configured]: https://github.com/golang/go/wiki/SettingGOPATH
[Releases]: https://github.com/dimaunx/armada/releases/
[kind release page]: https://github.com/kubernetes-sigs/kind/releases/tag/v0.5.0
[kind]: https://github.com/kubernetes-sigs/kind