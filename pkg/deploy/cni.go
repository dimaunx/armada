package deploy

import (
	"bytes"
	"text/template"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/dimaunx/armada/pkg/config"
	"github.com/dimaunx/armada/pkg/waiter"
	"github.com/gobuffalo/packr/v2"
	apiextclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

// Calico deploys calico cni
func Calico(cl *config.Cluster, kubeConfigFilePath string, box *packr.Box) error {

	kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return err
	}

	apiExtClientSet, err := apiextclientset.NewForConfig(kconfig)
	if err != nil {
		return err
	}

	calicoCrdFile, err := box.Resolve("tpl/calico-crd.yaml")
	if err != nil {
		return err
	}

	err = DeployCrdResources(cl, apiExtClientSet, calicoCrdFile.String())
	if err != nil {
		return err
	}

	calicoDeploymentTemplate, err := box.Resolve("tpl/calico-daemonset.yaml")
	if err != nil {
		return err
	}

	t, err := template.New("calico").Parse(calicoDeploymentTemplate.String())
	if err != nil {
		return err
	}

	var calicoDeploymentFile bytes.Buffer
	err = t.Execute(&calicoDeploymentFile, cl)
	if err != nil {
		return err
	}

	err = DeployResources(cl, clientSet, calicoDeploymentFile.String(), "Calico")
	if err != nil {
		return err
	}

	err = waiter.WaitForDaemonSet(cl, clientSet, "kube-system", "calico-node")
	if err != nil {
		return err
	}

	err = waiter.WaitForDeployment(cl, clientSet, "kube-system", "coredns")
	if err != nil {
		return err
	}
	return nil
}

// Weave deploys weave cni
func Weave(cl *config.Cluster, kubeConfigFilePath string, box *packr.Box) error {
	kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return err
	}

	weaveDeploymentTemplate, err := box.Resolve("tpl/weave-daemonset.yaml")
	if err != nil {
		return err
	}

	t, err := template.New("weave").Parse(weaveDeploymentTemplate.String())
	if err != nil {
		return err
	}

	var weaveDeploymentFile bytes.Buffer
	err = t.Execute(&weaveDeploymentFile, cl)
	if err != nil {
		return err
	}

	err = DeployResources(cl, clientSet, weaveDeploymentFile.String(), "Weave")
	if err != nil {
		return err
	}

	err = waiter.WaitForDaemonSet(cl, clientSet, "kube-system", "weave-net")
	if err != nil {
		return err
	}

	err = waiter.WaitForDeployment(cl, clientSet, "kube-system", "coredns")
	if err != nil {
		return err
	}
	return nil
}

// Flannel deploys flannel cni
func Flannel(cl *config.Cluster, kubeConfigFilePath string, box *packr.Box) error {
	kconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		return err
	}

	flannelDeploymentTemplate, err := box.Resolve("tpl/flannel-daemonset.yaml")
	if err != nil {
		return err
	}

	t, err := template.New("flannel").Parse(flannelDeploymentTemplate.String())
	if err != nil {
		return err
	}

	var flannelDeploymentFile bytes.Buffer
	err = t.Execute(&flannelDeploymentFile, cl)
	if err != nil {
		return err
	}

	err = DeployResources(cl, clientSet, flannelDeploymentFile.String(), "Flannel")
	if err != nil {
		return err
	}

	err = waiter.WaitForDaemonSet(cl, clientSet, "kube-system", "kube-flannel-ds-amd64")
	if err != nil {
		return err
	}

	err = waiter.WaitForDeployment(cl, clientSet, "kube-system", "coredns")
	if err != nil {
		return err
	}
	return nil
}
