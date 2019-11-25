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

	// KindConfigDir is a default kind config files destination directory
	KindConfigDir = "output/kind-clusters"

	// LocalKubeConfigDir is a default local workstation kubeconfig files destination directory
	LocalKubeConfigDir = "output/kind-config/local-dev"

	// LocalKubeConfigDir is a default  kubeconfig files destination directory if running inside container
	ContainerKubeConfigDir = "output/kind-config/container"

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
}

// CreateFlagpole is a list of cli flags for create clusters command
type CreateFlagpole struct {
	// ImageName is the node image used for cluster creation
	ImageName string

	// Wait is a time duration to wait until cluster is ready
	Wait time.Duration

	// Retain if you keep clusters running even if error occurs
	Retain bool

	// Weave if to install weave cni
	Weave bool

	// Flannel if to install flannel cni
	Flannel bool

	// Calico if to install calico cni
	Calico bool

	// Kindnet if to install kindnet default cni
	Kindnet bool

	// Debug if to enable debug log level
	Debug bool

	// Tiller if to install tiller
	Tiller bool

	// Overlap if to create clusters with overlapping cidrs
	Overlap bool

	// NumClusters is the number of clusters to create
	NumClusters int
}
