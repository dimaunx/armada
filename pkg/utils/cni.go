package utils

import (
	"bytes"
	"github.com/dimaunx/armada/pkg/config"
	"github.com/gobuffalo/packr/v2"
	"text/template"
)

// GenerateCalicoDeploymentFile generates calico deployment file from template
func GenerateCalicoDeploymentFile(cl *config.Cluster, box *packr.Box) (string, error) {
	calicoDeploymentTemplate, err := box.Resolve("tpl/calico-daemonset.yaml")
	if err != nil {
		return "", err
	}

	t, err := template.New("calico").Parse(calicoDeploymentTemplate.String())
	if err != nil {
		return "", err
	}

	var calicoDeploymentFile bytes.Buffer
	err = t.Execute(&calicoDeploymentFile, cl)
	if err != nil {
		return "", err
	}
	return calicoDeploymentFile.String(), nil
}

// GenerateFlannelDeploymentFile generates flannel deployment file from template
func GenerateFlannelDeploymentFile(cl *config.Cluster, box *packr.Box) (string, error) {
	flannelDeploymentTemplate, err := box.Resolve("tpl/flannel-daemonset.yaml")
	if err != nil {
		return "", err
	}

	t, err := template.New("flannel").Parse(flannelDeploymentTemplate.String())
	if err != nil {
		return "", err
	}

	var flannelDeploymentFile bytes.Buffer
	err = t.Execute(&flannelDeploymentFile, cl)
	if err != nil {
		return "", err
	}
	return flannelDeploymentFile.String(), nil
}

// GenerateWeaveDeploymentFile generates weave deployment file from template
func GenerateWeaveDeploymentFile(cl *config.Cluster, box *packr.Box) (string, error) {
	weaveDeploymentTemplate, err := box.Resolve("tpl/weave-daemonset.yaml")
	if err != nil {
		return "", err
	}

	t, err := template.New("weave").Parse(weaveDeploymentTemplate.String())
	if err != nil {
		return "", err
	}

	var weaveDeploymentFile bytes.Buffer
	err = t.Execute(&weaveDeploymentFile, cl)
	if err != nil {
		return "", err
	}
	return weaveDeploymentFile.String(), nil
}