package utils

import (
	"context"
	"github.com/dimaunx/armada/pkg/config"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/gobuffalo/packr/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	kind "sigs.k8s.io/kind/pkg/cluster"
	"strings"
	"text/template"
)

// GetMasterDockerIP gets control plain master docker internal ip
func GetMasterDockerIP(clName string) (string, error) {
	ctx := context.Background()
	dockerCli, err := dockerclient.NewEnvClient()
	if err != nil {
		return "", err
	}

	containerFilter := filters.NewArgs()
	containerFilter.Add("name", strings.Join([]string{clName, "control-plane"}, "-"))
	containers, err := dockerCli.ContainerList(ctx, dockertypes.ContainerListOptions{
		Filters: containerFilter,
		Limit:   1,
	})
	if err != nil {
		return "", err
	}
	return containers[0].NetworkSettings.Networks["bridge"].IPAddress, nil
}

// iterate func map for config template
func iterate(start, end int) (stream chan int) {
	stream = make(chan int)
	go func() {
		for i := start; i <= end; i++ {
			stream <- i
		}
		close(stream)
	}()
	return
}

// ClusterIsKnown returns bool if cluster exists
func ClusterIsKnown(clName string, provider *kind.Provider) (bool, error) {
	n, err := provider.ListNodes(clName)
	if err != nil {
		return false, err
	}
	if len(n) != 0 {
		return true, nil
	}
	return false, nil
}

// GenerateKindConfig creates kind config file and returns its path
func GenerateKindConfig(cl *config.Cluster, configDir string, box *packr.Box) (string, error) {
	kindConfigFileTemplate, err := box.Resolve("tpl/cluster-config.yaml")
	if err != nil {
		return "", err
	}

	t, err := template.New("config").Funcs(template.FuncMap{"iterate": iterate}).Parse(kindConfigFileTemplate.String())
	if err != nil {
		return "", err
	}

	kindConfigFilePath := filepath.Join(configDir, "kind-config-"+cl.Name+".yaml")
	f, err := os.Create(kindConfigFilePath)
	if err != nil {
		return "", err
	}

	err = t.Execute(f, cl)
	if err != nil {
		return "", err
	}

	if err := f.Close(); err != nil {
		return "", err
	}
	log.Debugf("Cluster config file for %s generated.", cl.Name)
	return kindConfigFilePath, nil
}
