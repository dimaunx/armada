package config

import "time"

// Default values
const (
	// ClusterNameBase is the default prefix for all cluster names
	ClusterNameBase = "cluster"

	// PodCidrBase the default starting pod cidr for all the clusters
	PodCidrBase = "10.0.0.0"

	// PodCidrMask is the default mask for pod subnet
	PodCidrMask = "/14"

	// ServiceCidrBase the default starting service cidr for all the clusters
	ServiceCidrBase = "100.0.0.0"

	// ServiceCidrMask is the default mask for service subnet
	ServiceCidrMask = "/16"

	// NumWorkers is the number of worker nodes per cluster
	NumWorkers = 2

	// KindLogsDir is a default kind log files destination directory
	KindLogsDir = "output/logs"

	// KindConfigDir is a default kind config files destination directory
	KindConfigDir = "output/kind-clusters"

	// LocalKubeConfigDir is a default local workstation kubeconfig files destination directory
	LocalKubeConfigDir = "output/kube-config/local-dev"

	// LocalKubeConfigDir is a default  kubeconfig files destination directory if running inside container
	ContainerKubeConfigDir = "output/kube-config/container"

	// WaitDurationResources is a default timeout for waiter functions
	WaitDurationResources = time.Duration(10) * time.Minute

	// KubeAdminAPIVersion is a default version used by in kind configs
	KubeAdminAPIVersion = "kubeadm.k8s.io/v1beta2"
)

// Cluster type
type Cluster struct {
	// Cni is a name of the cni that will be installed for a cluster
	Cni string

	// Name is a cluster name
	Name string

	// PodSubnet is pod subnet cidr and mask
	PodSubnet string

	// ServiceSubnet is a service subnet cidr and mask
	ServiceSubnet string

	// DNSDomain is cluster dns domain name
	DNSDomain string

	// // KubeAdminAPIVersion for each cluster
	KubeAdminAPIVersion string

	// NumWorkers is the number of worker nodes
	NumWorkers int

	// KubeConfigFilePath is the destination where kind will generate the original kubeconfig file
	KubeConfigFilePath string

	// Amount of time to wait for control plain to be ready
	WaitForReady time.Duration

	// Cluster image name
	NodeImageName string

	// Retain if to retain the cluster despite and error
	Retain bool

	// Tiller if to deploy a cluster with tiller
	Tiller bool
}
