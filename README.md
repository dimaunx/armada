# Armada

[![Build Status](https://travis-ci.com/dimaunx/armada.svg?branch=master)](https://travis-ci.com/dimaunx/armada)

## Prerequisites

- [go 1.12] with [$GOPATH configured]

## Get the latest version from [Releases] page.


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

The tool will create 2 clusters by default with kindnet.

```bash
cd ./bin
./armada create clusters
``` 

To create a higher number of clusters.

```bash
./armada create clusters -n 5
```

Create a total of three clusters, 2 with weave and 1 with flannel.

```bash
./armada create clusters -n 2 --weave
./armada create clusters -n 3 --flannel
```

Alternatively run in docker after **make docker-build**.

```bash
make docker-run ARGS="./armada create clusters -n 4 --flannel"
``` 

## Destroy clusters

```bash
armada destroy clusters
``` 

<!--links-->
[go 1.12]: https://blog.golang.org/go1.12
[$GOPATH configured]: https://github.com/golang/go/wiki/SettingGOPATH
[Releases]: https://github.com/dimaunx/armada/releases/