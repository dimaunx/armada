package constants

import "time"

// Default values
const (
	ClusterNameBase        = "cl"
	PodCidrBase            = "10.0.0.0"
	PodCidrMask            = "/14"
	ServiceCidrBase        = "100.0.0.0"
	ServiceCidrMask        = "/16"
	NumWorkers             = 2
	KindConfigDir          = "output/kind-clusters"
	LocalKubeConfigDir     = "output/kind-config/local-dev"
	ContainerKubeConfigDir = "output/kind-config/container"
	WaitDurationResources  = time.Duration(5) * time.Minute
)
